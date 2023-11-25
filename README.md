[![Go](https://github.com/ThatOneRuffian/tempgo/actions/workflows/go.yml/badge.svg)](https://github.com/ThatOneRuffian/tempgo/actions/workflows/go.yml)
# tempgo
A tempo utility written in golang to help with reverse engineering tempos and practicing a set tempo. Written with fyne as the GUI component.

In order to gain direct access to input devices, add your user to the input group (sudo usermod -aG input username) and restart the user session to apply the changes. Do not run this program as root. Alsa, the required sound component, will not allow root for security reasons.
