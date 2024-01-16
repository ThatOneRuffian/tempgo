[![Go](https://github.com/ThatOneRuffian/tempgo/actions/workflows/go.yml/badge.svg)](https://github.com/ThatOneRuffian/tempgo/actions/workflows/go.yml)
# tempgo
A tempo utility written in golang to help with reverse engineering tempos and practicing a set tempo. Written with fyne as the GUI component.

In order to gain direct access to input devices, add your user to the input group (sudo usermod -aG input username) and restart the user session to apply the changes. Do not run this program as root. Alsa, the required sound component, will not allow root for security reasons.

Zeros in dataset are ignored when calculating stats.

For best performance it is recommended to use low-latency/non-wireless input and listening equipment.

Thanks to Ludwig Peter Müller for allowing the use of the metronome sfx (under CC0 1.0 Universa).

Features:
- Compare input tempo against metronome
- Setup custom BPM and count
- Metronome-free input stats
- Select from various input devices
- Tray icon / background capture
- ARM64 Linux build
  
  ![image](https://github.com/ThatOneRuffian/tempgo/assets/13604240/97b72c94-a77f-4648-995e-5db8c5b9b878)

