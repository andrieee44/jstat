# JSTAT

[NAME](#NAME)
[SYNOPSIS](#SYNOPSIS)
[DESCRIPTION](#DESCRIPTION)
[CONFIGURATION](#CONFIGURATION)
[EXAMPLE OUTPUT](#EXAMPLE%20OUTPUT)
[UserHost (struct):](#UserHost%20(struct)%3A)
[Date (struct):](#Date%20(struct)%3A)
[Uptime (struct):](#Uptime%20(struct)%3A)
[Bat (map\[string\]struct):](#Bat%20(map%5Bstring%5Dstruct)%3A)
[Cpu (struct):](#Cpu%20(struct)%3A)
[Bri (struct):](#Bri%20(struct)%3A)
[Disk (map\[string\]struct):](#Disk%20(map%5Bstring%5Dstruct)%3A)
[Swap (struct):](#Swap%20(struct)%3A)
[Ram (struct):](#Ram%20(struct)%3A)
[Vol (struct):](#Vol%20(struct)%3A)
[Music (struct):](#Music%20(struct)%3A)
[Internet (struct):](#Internet%20(struct)%3A)
[Ethernet (struct):](#Ethernet%20(struct)%3A)
[Hyprland (struct):](#Hyprland%20(struct)%3A)
[Bluetooth (struct):](#Bluetooth%20(struct):)
[SEE ALSO](#SEE%20ALSO)
[AUTHOR](#AUTHOR)

------------------------------------------------------------------------

## NAME <span id="NAME"></span>

jstat − system statistics in JSON

## SYNOPSIS <span id="SYNOPSIS"></span>

**jstat**

## DESCRIPTION <span id="DESCRIPTION"></span>

**jstat** is a system statistics aggregator, originally designed to work
with *eww*.

## CONFIGURATION <span id="CONFIGURATION"></span>

Configuration is done by altering **cmd/jstat/config.go**. See
**https://github.com/andrieee44/jstat** for more information.

## EXAMPLE OUTPUT <span id="EXAMPLE OUTPUT"></span>

### UserHost (struct): <span id="UserHost (struct):"></span>

**UID (string):**

The user’s ID.

**GID (string):**

The user’s group ID.

**Name (string):**

The user’s username.

**Host (string):**

The hostname of the system.

### Date (struct): <span id="Date (struct):"></span>

**Date (string):**

The current time.

**Icon (string):**

The hour icon.

### Uptime (struct): <span id="Uptime (struct):"></span>

**Hours (int):**

The amount of uptime hours.

**Minutes (int):**

The amount of uptime minutes.

**Seconds (int):**

The amount of uptime seconds.

### Bat (map\[string\]struct): <span id="Bat (map[string]struct):"></span>

**Status (string):**

The current status of the battery. Possible values are: **"Charging"**,
**"Discharging"**, **"Full"**, **"Not Charging"**, **"Unknown"**.

**Capacity (int):**

The battery capacity in percentage.

**Icon (string):**

The battery capacity icon.

### Cpu (struct): <span id="Cpu (struct):"></span>

**Cores (map\[int\]struct):**

The list of CPU cores in the system.

**Freq (int):**

The frequency of the CPU core.

**Usage (float64):**

The CPU core usage in percentage.

**AvgUsage (float64):**

The average usage of all CPU cores.

**Icon (string):**

The average usage icon of all CPU cores.

### Bri (struct): <span id="Bri (struct):"></span>

**Perc (float64):**

The screen brightness percentage.

**Icon (string):**

The screen brightness percentage icon.

### Disk (map\[string\]struct): <span id="Disk (map[string]struct):"></span>

**Total (int):**

The total disk space in *path*.

**Free (int):**

The free disk space available in *path*.

**Used (int):**

The used disk space in *path*.

**UsedPerc (float64):**

The used disk space percentage in *path*.

**Icon (string):**

The used disk space percentage icon in *path*.

### Swap (struct): <span id="Swap (struct):"></span>

**Total (int):**

The total swap space.

**Free (int):**

The free swap space.

**Used** (int):

The used swap space.

**UsedPerc (float64):**

The used swap space percentage.

**Icon (string):**

The used swap space percentage icon.

### Ram (struct): <span id="Ram (struct):"></span>

**Total (int):**

The total RAM in the system.

**Free (int):**

The free RAM in the system.

**Available (int):**

The available RAM in the system.

**Used (int):**

The used RAM in the system.

**UsedPerc (float64):**

The used RAM percentage in the system.

**Icon (string):**

The used RAM percentage icon.

### Vol (struct): <span id="Vol (struct):"></span>

**Perc (float64):**

The volume loudness percentage.

**Mute (bool):**

Whether the volume is muted.

**Icon (string):**

The volume loudness percentage icon.

### Music (struct): <span id="Music (struct):"></span>

**Song (string):**

The current song loaded in *mpd*(1).

**State (string):**

The state of the music player of *mpd*(1). Possible values are
**"play"**, **"pause"**, **"stop"**.

**Scroll (int):**

The index offset for text scrolling effect.

**Limit (int):**

The scroll limit.

### Internet (struct): <span id="Internet (struct):"></span>

**Internets (map\[string\]struct):**

The list of available system wifi interfaces.

**Name (string):**

The SSID of the wifi.

**Icon (string):**

The signal strength percentage of the wifi icon.

**Powered (bool):**

Whether the system wifi interface is powered on.

**Scanning (bool):**

Whether the system wifi interface is scanning for available wifis.

**Scroll (int):**

The index offset for text scrolling effect.

**Strength (float64):**

The signal strength percentage of the wifi.

**Limit (int):**

The scroll limit.

### Ethernet (struct): <span id="Ethernet (struct):"></span>

**Ethernets (map\[string\]struct):**

The list of available system ethernet interfaces.

**Powered (bool):**

Whether the system ethernet interface is powered on.

**Scroll (int):**

The index offset for text scrolling effect.

**Limit (int):**

The scroll limit.

### Hyprland (struct): <span id="Hyprland (struct):"></span>

**Window (string):**

The name of the current active window in Hyprland.

**Monitors (map\[int\]struct):**

The list of available monitors in the system.

**Name (string):**

The name of the monitor.

**Workspaces (map\[int\]string):**

The name of the nth workspace of the monitor.

**ActiveMonitor (int):**

The current active monitor in Hyprland.

**ActiveWorkspace (int):**

The current active workspace in Hyprland.

**Scroll (int):**

The index offset for text scrolling effect.

**Limit (int):**

The scroll limit.

### Bluetooth (struct): <span id="Bluetooth (struct):"></span>

**Adapter (map\[string\]struct):**

The bluetooth adapters on the system.

**Name (string):**

The bluetooth adapter interface name.

**Scroll (int):**

The index offset for text scrolling effect.

**Powered (bool):**

Whether the bluetooth adapter is powered on.

**Discovering (bool):**

Whether the bluetooth adapter is discovering other available devices.

**Devices (map\[string\]struct):**

The devices known to the bluetooth adapter.

**Name (string):**

The name of the device.

**Icon (string):**

The battery capacity icon.

**Battery (int):**

The battery capacity in percentage.

**Scroll (int):**

The index offset for text scrolling effect.

**Connected (bool):**

Whether the device is currently connected to the bluetooth adapter.

**Limit (int):**

The scroll limit.

## SEE ALSO <span id="SEE ALSO"></span>

*mpd*(1).

## AUTHOR <span id="AUTHOR"></span>

andrieee44 (andrieee44@gmail.com)

------------------------------------------------------------------------
