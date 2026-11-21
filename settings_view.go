package subsonic

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/waygu-music-plugin/sdk/desktop"

	"github.com/waygu-music-plugin/sdk/pluginapi"
)

func (h *Home) desktopSettings(c *ui.Context) {
	if message := h.host.TakeToast(); message != "" {
		c.Toast(message)
	}
	p := paletteFor(c.Theme())
	state := h.host.DesktopSettings()
	s := &state.Preferences
	apply := func() { h.host.ApplyDesktopPreferences(*s) }
	group := func(title string, rows func()) {
		ui.Text(c, title).FontSize(12).FontWeight(600).TextColor(p.muted).Margin(14, 4, 4, 4)
		ui.Column(c).Radius(12).Background(p.bg).Border(1, p.hairline).Dividers(1, p.hairline).Children(rows)
	}
	row := func(key, label, desc string, control func()) {
		ui.Row(c.Key(key)).Padding(12, 16).Gap(16).AlignItems(ui.Center).MinHeight(48).Children(func() {
			ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(3).Children(func() {
				ui.Text(c, label).FontSize(13.5).FontWeight(500)
				if desc != "" {
					ui.Text(c, desc).FontSize(12).TextColor(p.muted)
				}
			})
			control()
		})
	}
	h.audioSettings(c, s, group, row, apply)
	group("缓存", func() {
		row("shared-cache", "歌曲缓存", cacheUsage(state.CacheUsed), func() {
			labels := make([]string, len(cacheLimits))
			chosen := ""
			for i, v := range cacheLimits {
				labels[i] = v.label
				if v.mb == s.CacheMB {
					chosen = v.label
				}
			}
			if ui.Select(c.Key("shared-cache-select"), &chosen, labels).Label("歌曲缓存").Width(150).Changed() {
				for _, v := range cacheLimits {
					if v.label == chosen {
						s.CacheMB = v.mb
					}
				}
				apply()
			}
		})
		clearDescription := "删除本机保存的歌曲"
		if state.CacheTask.Message != "" {
			clearDescription = state.CacheTask.Message
		}
		row("shared-cache-clear", "清除缓存", clearDescription, func() {
			if ui.Button(c.Key("shared-cache-clear-button"), "清除").Disabled(state.CacheUsed <= 0 || state.CacheTask.Running()).Clicked() {
				h.host.ClearAudioCache()
			}
		})
	})
	group("外观", func() {
		row("shared-appearance", "外观", "跟随系统或固定主题", func() {
			modes := []string{"system", "light", "dark"}
			i := 0
			for k, v := range modes {
				if v == s.Appearance {
					i = k
				}
			}
			if segmented(c.Key("shared-appearance-segment"), &i, "跟随系统", "浅色", "深色") {
				s.Appearance = modes[i]
				apply()
			}
		})
	})
	if runtime.GOOS == "darwin" {
		group("应用图标", func() {
			row("shared-app-icon", "应用图标", "显示在菜单栏时使用的图标", func() {
				labels := make([]string, len(appIcons))
				chosen := appIcons[0].label
				for i, ic := range appIcons {
					labels[i] = ic.label
					if ic.id == s.AppIcon {
						chosen = ic.label
					}
				}
				if ui.Select(c.Key("shared-app-icon-select"), &chosen, labels).Label("应用图标").Width(150).Changed() {
					for _, ic := range appIcons {
						if ic.label == chosen {
							s.AppIcon = ic.id
						}
					}
					apply()
				}
			})
		})
	}
	if state.DesktopLyricsColorAvailable || state.StatusLyricsAvailable || state.LyricsFontAvailable {
		group("歌词", func() {
			if state.LyricsFontAvailable {
				row("shared-lyrics-font-family", "歌词字体", "", func() {
					desktop.LyricsFontFamilyControl(c.Key("lyrics-font-family"), h.host, state)
				})
			}
			if state.DesktopLyricsColorAvailable {
				row("shared-lyrics-color", "桌面歌词颜色", "", func() {
					desktop.LyricsColorControl(c.Key("desktop-lyrics-color"), h.host, state, h.desktopLyricsPreset(h.host.DesktopLyricsState().Color))
				})
			}
			if state.StatusLyricsAvailable {
				row("shared-status-lyrics", "任务栏歌词", "在系统任务栏或菜单栏显示歌词", func() {
					desktop.StatusLyricsControl(c.Key("status-lyrics"), h.host, state)
				})
			}
			if state.TaskbarLyricsFontSize > 0 {
				row("shared-taskbar-lyrics-font", "任务栏歌词字号", "", func() {
					desktop.TaskbarLyricsFontControl(c.Key("taskbar-lyrics-font"), h.host, state)
				})
			}
		})
	}
	if state.NotchAvailable {
		group("刘海播放器", func() {
			row("shared-notch", "显示刘海播放器", "在屏幕顶部显示迷你播放器", func() {
				if ui.Switch(c.Key("shared-notch-switch"), &s.Notch).Label("显示刘海播放器").Changed() {
					apply()
				}
			})
			row("shared-notch-activity", "播放时显示实时活动", "显示封面与波形", func() {
				if ui.Switch(c.Key("shared-notch-activity-switch"), &s.NotchActivity).Label("播放时显示实时活动").Disabled(!s.Notch).Changed() {
					apply()
				}
			})
			row("shared-notch-virtual", "在无刘海的屏幕上常驻显示", "关闭时，将鼠标移至屏幕顶部中央即可唤出", func() {
				if ui.Switch(c.Key("shared-notch-virtual-switch"), &s.NotchVirtual).Label("在无刘海的屏幕上常驻显示").Disabled(!s.Notch).Changed() {
					apply()
				}
			})
		})
	}
	group("实验性功能", func() {
		row("shared-smoke", "进度条烟雾特效", "播放时进度条化作随音乐起伏的光点", func() {
			if ui.Switch(c.Key("shared-smoke-switch"), &s.Smoke).Label("进度条烟雾特效").Changed() {
				apply()
			}
		})
		row("shared-fog", "烟雾背景", "播放页背景随封面色调缓缓升起", func() {
			if ui.Switch(c.Key("shared-fog-switch"), &s.Fog).Label("烟雾背景").Changed() {
				apply()
			}
		})
	})
	if state.Updates {
		group("更新", func() {
			row("shared-update-checks", "自动检查更新", "每天检查一次是否有新版本", func() {
				on := state.UpdateChecks
				if ui.Switch(c.Key("shared-update-checks-switch"), &on).Label("自动检查更新").Changed() {
					h.host.SetUpdateChecks(on)
					state.UpdateChecks = on
				}
			})
			row("shared-update-downloads", "自动下载并安装", "在后台下载新版本，下次打开时生效", func() {
				on := state.UpdateDownloads
				if ui.Switch(c.Key("shared-update-downloads-switch"), &on).Label("自动下载并安装").Disabled(!state.UpdateChecks).Changed() {
					h.host.SetUpdateDownloads(on)
				}
			})
			row("shared-version", "当前版本 "+mygo.App.Version(), "", func() {
				if ui.Button(c.Key("shared-check-updates"), "检查更新").Clicked() {
					h.host.CheckForUpdates()
				}
			})
		})
	} else {
		group("关于", func() {
			row("shared-version", "当前版本 "+mygo.App.Version(), "", func() {})
		})
	}
}

func (h *Home) audioSettings(c *ui.Context, s *pluginapi.DesktopPreferences, group func(string, func()), row func(string, string, string, func()), apply func()) {
	status := h.host.DecoderStatus()
	group("高级音频与硬件解码", func() {
		row("hifi-mpv-toggle", "高级解码引擎 (MPV)", "基于 MPV 的发烧级解码引擎，支持 DSD (DSF/DFF)、APE、ALAC、AAC、Hi-Res FLAC 全格式直接硬件输出", func() {
			if ui.Switch(c.Key("use-mpv"), &s.UseMPV).Label("高级解码引擎").Disabled(!status.Ready).Changed() {
				apply()
				status = h.host.DecoderStatus()
			}
		})

		var statusDesc string
		switch {
		case status.Downloading:
			statusDesc = fmt.Sprintf("正在下载 libmpv (%s / %s)...", megabytes(status.Done), megabytes(status.Total))
		case status.DeleteTask.Running():
			statusDesc = "正在删除解码组件…"
		case status.LatestTask().Message != "":
			statusDesc = status.LatestTask().Message
		case status.Downloaded:
			statusDesc = "组件已就绪 (独立组件已缓存)"
		case status.SystemPath != "":
			statusDesc = "系统已安装 libmpv (" + filepath.Base(status.SystemPath) + ")，亦可下载独立组件"
		default:
			statusDesc = "未下载组件，下载后即可开启高级解码引擎"
		}

		row("hifi-mpv-status", "解码组件状态", statusDesc, func() {
			switch {
			case status.DeleteTask.Running():
				ui.Button(c.Key("mpv-dl"), "删除组件").Disabled(true)
			case status.Downloading:
				if ui.Button(c.Key("mpv-dl"), "取消").Clicked() {
					h.host.CancelDecoderDownload()
				}
			case status.Downloaded:
				if ui.Button(c.Key("mpv-dl"), "删除组件").Clicked() {
					h.host.DeleteDecoder()
				}
			default:
				if ui.Button(c.Key("mpv-dl"), "下载组件").Clicked() {
					h.host.StartDecoderDownload()
				}
			}
		})

		if s.UseMPV && status.Ready {
			devices := status.Devices
			if len(devices) > 0 {
				var devLabels []string
				chosenDevLabel := "系统默认 (自动选择)"
				devNameToLabel := make(map[string]string)
				labelToDevName := make(map[string]string)
				labelCounts := make(map[string]int)

				for _, d := range devices {
					lbl := d.Description
					if lbl == "" {
						lbl = d.Name
					}
					if d.Name == "auto" {
						lbl = "系统默认 (自动选择)"
					}
					labelCounts[lbl]++
				}

				for _, d := range devices {
					lbl := d.Description
					if lbl == "" {
						lbl = d.Name
					}
					if d.Name == "auto" {
						lbl = "系统默认 (自动选择)"
					} else if labelCounts[lbl] > 1 {
						lbl = fmt.Sprintf("%s (%s)", lbl, d.Name)
					}
					devLabels = append(devLabels, lbl)
					devNameToLabel[d.Name] = lbl
					labelToDevName[lbl] = d.Name
				}

				if curLbl, ok := devNameToLabel[s.MPVAudioDevice]; ok {
					chosenDevLabel = curLbl
				} else if s.MPVAudioDevice == "" || s.MPVAudioDevice == "auto" {
					chosenDevLabel = "系统默认 (自动选择)"
				}

				row("hifi-mpv-device", "音频输出设备", "独立选择音频路由设备 (如耳机、板载声卡或外置 USB DAC/耳放)", func() {
					if ui.Select(c.Key("mpv-audio-device"), &chosenDevLabel, devLabels).Label("音频输出设备").Width(240).Changed() {
						if targetName, ok := labelToDevName[chosenDevLabel]; ok {
							s.MPVAudioDevice = targetName
							apply()
						}
					}
				})
			}

			row("hifi-mpv-exclusive", "独占音频输出 (Exclusive Mode)", "绕过系统音频混音器，提供无损 Bit-perfect 直通原音输出", func() {
				if ui.Switch(c.Key("mpv-exclusive"), &s.MPVExclusive).Label("独占音频输出").Changed() {
					apply()
				}
			})
			row("hifi-mpv-gapless", "无缝切歌 (Gapless Playback)", "连续曲目切换时保持音频流连续，消除曲目间隙缓冲与爆音", func() {
				if ui.Switch(c.Key("mpv-gapless"), &s.MPVGapless).Label("无缝切歌").Changed() {
					apply()
				}
			})

			row("hifi-mpv-eq-info", "图形均衡器 (ISO 10/15 段)", "基于专业音频滤镜的图形调音台，已整合至底部播控条右侧，可随时快捷唤出", func() {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, iconEqualizer).Size(14, 14).TextColor(c.Theme().TextMuted)
					ui.Text(c, "播控条已启用快捷入口").FontSize(12).TextColor(c.Theme().TextMuted)
				})
			})
		}
	})
}

func cacheUsage(used int64) string {
	switch n := used; {
	case n < 0:
		return "正在计算已用空间…"
	case n >= 1<<30:
		return fmt.Sprintf("已用 %.1f GB", float64(n)/(1<<30))
	default:
		return fmt.Sprintf("已用 %d MB", n>>20)
	}
}

func megabytes(n int64) string {
	if n >= 1<<30 {
		return fmt.Sprintf("%.2f GB", float64(n)/(1<<30))
	}
	return fmt.Sprintf("%d MB", n>>20)
}

var cacheLimits = []struct {
	mb    int
	label string
}{
	{0, "关闭"},
	{1024, "1 GB"},
	{2048, "2 GB"},
	{5120, "5 GB"},
	{10240, "10 GB"},
}

var appIcons = []struct{ id, label string }{
	{"dock-and-menubar", "程序坞与菜单栏"},
	{"dock", "仅程序坞"},
	{"menubar", "仅菜单栏"},
}
