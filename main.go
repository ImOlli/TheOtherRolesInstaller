//go:generate goversioninfo -icon=icon.ico
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ncruces/zenity"
)

func main() {
	action := zenity.Question("Welcome to the Among Us TheOtherRoles Installer. Please make sure to have the latest version of Among Us Installed.",
		zenity.Title("AmongUs TheOtherRoles Installer"),
		zenity.OKLabel("Install"),
		zenity.CancelLabel("Cancel"),
	)

	if action != nil {
		// Close installer
		os.Exit(0)
	}

	dlg, err := zenity.Progress(
		zenity.Title("Installing TheOtherRoles Mod"),
		zenity.NoCancel(),
		zenity.OKLabel("Cancel"))
	if err != nil {
		return
	}

	defer dlg.Close()

	installationPath := ""

	dlg.Text("Searching for AmongUs installation...")
	dlg.Value(1)

	// Load all Harddrives
	drives := getDrives()

	// Check on each drive if it contains AmongUs installation
	for drive := range drives {
		fmt.Println("Searching for AmongUs on drive " + drives[drive])

		if isValidAmongUsLocation(drives[drive] + ":\\Program Files (x86)\\Steam\\steamapps\\common\\Among Us") {
			installationPath = drives[drive] + ":\\Program Files (x86)\\Steam\\steamapps\\common\\Among Us"
			break
		}
	}

	dlg.Value(10)

	// No Among Us Installation found
	if installationPath == "" {
		fmt.Println("No AmongUs installation found")
		dlg.Text("Waiting for user input...")

		installationPath = selectAmongUsInstallationLocation()
	}

	fmt.Println("AmongUs installation path: " + installationPath)
	dlg.Value(20)

	dlg.Text("Downloading TheOtherRoles Mod...")
	fmt.Println("Downloading TheOtherRoles Mod...")

	// Creating temp directory
	tempFolder, err := os.MkdirTemp("", "amongus-theotherroles-installer")
	fmt.Println("Temp folder created: " + tempFolder)

	if err != nil {
		fmt.Println("Error creating temp folder: " + err.Error())
		showErrorAndClose()
		return
	}

	dlg.Value(30)

	// Download TheOtherRoles Mod
	// TODO Automatically download the newest version
	fileUrl := "https://github.com/Eisbison/TheOtherRoles/releases/download/v3.3.3/TheOtherRoles.zip"
	err = DownloadFile(filepath.Join(tempFolder, "TheOtherRoles.zip"), fileUrl)
	if err != nil {
		fmt.Println("Error while downloading file: " + err.Error())
		showErrorAndClose()
		return
	}

	dlg.Value(60)
	dlg.Text("Extracting TheOtherRoles Mod...")
	fmt.Println("Extracting TheOtherRoles Mod...")

	// Extract TheOtherRoles Mod
	_, err = unzip(filepath.Join(tempFolder, "TheOtherRoles.zip"), filepath.Join(tempFolder, "TheOtherRoles"))

	if err != nil {
		fmt.Println("Error while unziping files: " + err.Error())
		showErrorAndClose()
		return
	}
	dlg.Value(70)

	// Check if TheOtherRoles Installation alredy exists
	fmt.Println("Checking if TheOtherRoles Installation already exists...")
	dlg.Text("Checking if TheOtherRoles Installation already exists...")
	theOtherRolesInstallation := filepath.Dir(installationPath)
	theOtherRolesInstallation = filepath.Join(theOtherRolesInstallation, "Among Us TheOtherRoles v3.3.3")

	if isValidAmongUsLocation(theOtherRolesInstallation) {
		fmt.Println("Installation already exists. Asking user to overwrite it")
		// Ask user if he wants to overwrite the existing installation
		action := zenity.Question("TheOtherRoles Installation already exists. Do you want to overwrite it?", zenity.Title("Installation already exists"), zenity.OKLabel("Yes"), zenity.CancelLabel("No"))

		if action != nil {
			// User doesn't want to overwrite the existing installation
			fmt.Println("User doesn't want to overwrite the existing installation. Closing the installer")
			showErrorAndClose()
			return
		}

		// Delete TheOtherRoles Installation
		err = os.RemoveAll(theOtherRolesInstallation)
		if err != nil {
			fmt.Println("Error while deleting old Installation: " + err.Error())
			showErrorAndClose()
			return
		}
	}

	// Copying AmongUs Installation
	dlg.Text("Creating Among Us Installation")
	fmt.Println("Creating Among Us Installation")
	// TODO Check if permission is okay?
	os.Mkdir(theOtherRolesInstallation, 0777)
	copy(installationPath, theOtherRolesInstallation)

	// Copy TheOtherRoles Mod to AmongUs installation
	dlg.Text("Coping mod to installation...")
	fmt.Println("Coping mod to installation...")
	copy(filepath.Join(tempFolder, "TheOtherRoles"), theOtherRolesInstallation)

	// Clean up
	dlg.Text("Cleaning up...")
	fmt.Println("Cleaning up...")
	err = os.RemoveAll(tempFolder)

	if err != nil {
		fmt.Println("Error while deleting temp folder: " + err.Error())
	}

	// TODO Create shortcut to TheOtherRoles Mod Installation

	action = zenity.Question("TheOtherRoles Mod successfully installed. Enjoy!", zenity.OKLabel("Open Among Us"), zenity.CancelLabel("Close"))

	if action == nil {
		// Open AmongUs
		fmt.Println("Opening AmongUs")

	} else {
		// Close installer
		fmt.Println("Installation done!")
		os.Exit(1)
	}

}

// TODO other ways to copy the files
func copy(source, destination string) error {
	var err error = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			var data, err1 = ioutil.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})
	return err
}

func unzip(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func selectAmongUsInstallationLocation() string {
	action := zenity.Question("No AmongUs installation found! Would you like to select the installation by your self?",
		zenity.Title("No AmongUs installation found!"),
		zenity.OKLabel("Select"),
		zenity.CancelLabel("Cancel"))

	if action != nil {
		// User canceled installation
		os.Exit(0)
	}

	// Select AmongUs installation folder
	path, error := zenity.SelectFile(zenity.Title("Select AmongUs Installation Location"), zenity.Directory())

	// User selected no path
	if error != nil {
		fmt.Println("Error while selecting AmongUs installation folder")
		// Reopen dialog to ask for installation path
		return selectAmongUsInstallationLocation()
	}

	fmt.Println("User selected following installation path: " + path)

	if !isValidAmongUsLocation(path) {
		// The selected path is not a valid AmongUs installation path
		// Reopen dialog to ask for installation path
		return selectAmongUsInstallationLocation()
	}

	// Path is valid; returing
	return path
}

func showErrorAndClose() {
	zenity.Error("An error occured while installing TheOtherRoles Mod. Please try again later.")
	os.Exit(0)
}

func isValidAmongUsLocation(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else {
		return false
	}
}

func getDrives() (drives []string) {
	kernel32, _ := syscall.LoadLibrary("kernel32.dll")
	getLogicalDrivesHandle, _ := syscall.GetProcAddress(kernel32, "GetLogicalDrives")

	if ret, _, callErr := syscall.Syscall(uintptr(getLogicalDrivesHandle), 0, 0, 0, 0); callErr != 0 {
		// handle error
	} else {
		return bitsToDrives(uint32(ret))
	}

	return
}

func bitsToDrives(bitMap uint32) (drives []string) {
	availableDrives := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	for i := range availableDrives {
		if bitMap&1 == 1 {
			drives = append(drives, availableDrives[i])
		}
		bitMap >>= 1
	}

	return
}
