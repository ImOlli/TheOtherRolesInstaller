# TheOtherRoles Installer
This is a simple tool to install [TheOtherRoles Mod](https://github.com/Eisbison/TheOtherRoles) for Among Us. 

Currently only works for Windows!

## How to build
Execute go generate to generate the version info for the executable
> go generate

Then run this to build the executeable
> go build -ldflags -H=windowsgui

## Usage
After executing the installer just follow the steps of the dialogs. The Installer first tries to search for your among us installation. If it can't find your installation it will ask you to select it by your self. Then it will download the latest release of [TheOtherRoles Mod](https://github.com/Eisbison/TheOtherRoles) and creating a copy of your current among us installation and install the mod to it.