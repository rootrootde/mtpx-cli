package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	mtpx "github.com/ganeshrvel/go-mtpx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mtpx-cli <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  list <remote_path>                  List files at remote path")
		fmt.Println("  download <remote> <local_dir>       Download a file into target directory")
		fmt.Println("  upload <local_file> <remote_dir>    Upload a file into remote directory")
		fmt.Println("  delete <remote_path> [...]          Delete one or more files by remote path")
		fmt.Println("  stat <remote_path>                  Check if a file exists and print its size")
		fmt.Println("  device-info                         Show basic device information")
		fmt.Println("  storage-info                        Show storage-related information")
		os.Exit(1)
	}

	cmd := os.Args[1]

	dev, err := mtpx.Initialize(mtpx.Init{})
	if err != nil {
		log.Fatal(err)
	}
	defer mtpx.Dispose(dev)

	storages, err := mtpx.FetchStorages(dev)
	if err != nil || len(storages) == 0 {
		log.Fatal("no storage found")
	}
	sid := storages[0].Sid

	switch cmd {
	case "list":
		if len(os.Args) < 3 {
			log.Fatal("list requires remote path")
		}
		_, _, _, err := mtpx.Walk(dev, sid, os.Args[2], true, true, false,
			func(objectId uint32, fi *mtpx.FileInfo, err error) error {
				if err == nil {
					b, _ := json.Marshal(map[string]interface{}{
						"path": fi.FullPath,
						"size": fi.Size,
					})
					fmt.Println(string(b))
				}
				return nil
			})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("MTPX_LIST_DONE")

	case "download":
		if len(os.Args) < 4 {
			log.Fatal("download requires remote path and local target dir")
		}
		targetDir := ""
		targetDir, err = filepath.Abs(os.Args[3])
		if err != nil {
			log.Fatalf("invalid target path: %v", err)
		}
		printedDone := false // <-- define outside the callback!
		_, _, err := mtpx.DownloadFiles(dev, sid, []string{os.Args[2]}, targetDir, false,
			func(fi *mtpx.FileInfo, err error) error { return nil },
			func(pi *mtpx.ProgressInfo, err error) error {
				if pi.ActiveFileSize.Progress < 100.0 {
					b, _ := json.Marshal(map[string]interface{}{
						"file":     pi.FileInfo.Name,
						"progress": pi.ActiveFileSize.Progress,
					})
					fmt.Println(string(b))
				} else if pi.ActiveFileSize.Progress == 100.0 && !printedDone {
					// Always print a 100% progress message first
					b, _ := json.Marshal(map[string]interface{}{
						"file":     pi.FileInfo.Name,
						"progress": 100.0,
					})
					fmt.Println(string(b))
					printedDone = true
					// Then print the summary
					sourcePath := pi.FileInfo.FullPath
					targetPath := filepath.Join(targetDir, pi.FileInfo.Name)
					b2, _ := json.Marshal(map[string]string{
						"source": sourcePath,
						"target": targetPath,
					})
					fmt.Println(string(b2))
				}
				return nil
			})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("MTPX_DOWNLOAD_DONE")

	case "upload":
		if len(os.Args) < 4 {
			log.Fatal("upload requires local file and remote target dir")
		}
		localFile, err := filepath.Abs(os.Args[2])
		if err != nil {
			log.Fatalf("invalid local file path: %v", err)
		}
		printedDone := false // guard variable to prevent duplicate output
		_, _, _, err = mtpx.UploadFiles(dev, sid, []string{localFile}, os.Args[3], false,
			func(fi *os.FileInfo, path string, err error) error { return nil },
			func(pi *mtpx.ProgressInfo, err error) error {
				if pi.ActiveFileSize.Progress < 100.0 {
					b, _ := json.Marshal(map[string]interface{}{
						"file":     pi.FileInfo.Name,
						"progress": pi.ActiveFileSize.Progress,
					})
					fmt.Println(string(b))
				} else if pi.ActiveFileSize.Progress == 100.0 && !printedDone {
					// Always print a 100% progress message first
					b, _ := json.Marshal(map[string]interface{}{
						"file":     pi.FileInfo.Name,
						"progress": 100.0,
					})
					fmt.Println(string(b))
					printedDone = true
					// Then print the summary
					sourcePath := localFile
					targetPath := filepath.Join(os.Args[3], pi.FileInfo.Name)
					b2, _ := json.Marshal(map[string]string{
						"source": sourcePath,
						"target": targetPath,
					})
					fmt.Println(string(b2))
				}
				return nil
			})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("MTPX_UPLOAD_DONE")

	case "delete":
		if len(os.Args) < 3 {
			log.Fatal("delete requires at least one remote path")
		}

		var props []mtpx.FileProp
		for _, path := range os.Args[2:] {
			props = append(props, mtpx.FileProp{FullPath: path})
		}

		err = mtpx.DeleteFile(dev, sid, props)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("MTPX_DELETE_DONE")

	case "stat":
		if len(os.Args) < 3 {
			log.Fatal("stat requires a remote path")
		}
		props := []mtpx.FileProp{{FullPath: os.Args[2]}}
		results, err := mtpx.FileExists(dev, sid, props)
		if err != nil {
			log.Fatal(err)
		}
		info := results[0]
		if info.Exists {
			fi := info.FileInfo
			fmt.Printf("STAT\t%s\t%d\t%s\n", fi.FullPath, fi.Size, humanReadableSize(fi.Size))
		} else {
			fmt.Println("NOT_FOUND")
		}
		fmt.Println("MTPX_STAT_DONE")

	case "device-info":
		info, err := mtpx.FetchDeviceInfo(dev)
		if err != nil {
			log.Fatal(err)
		}
		b, _ := json.Marshal(info)
		fmt.Println(string(b))
		fmt.Println("MTPX_DEVICE_INFO_DONE")

	case "storage-info":
		storages, err := mtpx.FetchStorages(dev)
		if err != nil {
			log.Fatal("failed to fetch storage info:", err)
		}
		b, _ := json.Marshal(storages)
		fmt.Println(string(b))
		fmt.Println("MTPX_STORAGE_INFO_DONE")


	default:
		log.Fatal("unknown command")
	}
}

func humanReadableSize(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB"}
	size := float64(bytes)
	i := 0
	for size >= 1024 && i < len(units)-1 {
		size /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", size, units[i])
}
// describeStorageType returns a string description for a storage type code.
func describeStorageType(code uint16) string {
	switch code {
	case 0x0000:
		return "Undefined"
	case 0x0001:
		return "Fixed ROM"
	case 0x0002:
		return "Removable ROM"
	case 0x0003:
		return "Fixed RAM"
	case 0x0004:
		return "Removable RAM"
	default:
		return "Unknown"
	}
}

// describeFilesystemType returns a string description for a filesystem type code.
func describeFilesystemType(code uint16) string {
	switch code {
	case 0x0000:
		return "Undefined"
	case 0x0001:
		return "Generic Flat"
	case 0x0002:
		return "Generic Hierarchical"
	case 0x0003:
		return "DCF"
	default:
		return "Unknown"
	}
}

// describeAccessCapability returns a string description for an access capability code.
func describeAccessCapability(code uint16) string {
	switch code {
	case 0x0000:
		return "Read-Write"
	case 0x0001:
		return "Read-Only (with Object Deletion)"
	case 0x0002:
		return "Read-Only"
	default:
		return "Unknown"
	}
}