package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	mtpx "github.com/ganeshrvel/go-mtpx"
	"github.com/ganeshrvel/go-mtpfs/mtp"
)

// CLI represents the command line interface
type CLI struct {
	device  *mtp.Device
	storage uint32
}

// ProgressHandler manages progress output for transfers
type ProgressHandler struct {
	printedDone bool
	targetDir   string
	sourcePath  string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cli, err := newCLI()
	if err != nil {
		log.Fatal(err)
	}
	defer mtpx.Dispose(cli.device)

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "list":
		err = cli.handleList(args)
	case "download":
		err = cli.handleDownload(args)
	case "upload":
		err = cli.handleUpload(args)
	case "delete":
		err = cli.handleDelete(args)
	case "stat":
		err = cli.handleStat(args)
	case "device-info":
		err = cli.handleDeviceInfo(args)
	case "storage-info":
		err = cli.handleStorageInfo(args)
	default:
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func newCLI() (*CLI, error) {
	dev, err := mtpx.Initialize(mtpx.Init{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MTP: %w", err)
	}

	storages, err := mtpx.FetchStorages(dev)
	if err != nil || len(storages) == 0 {
		mtpx.Dispose(dev)
		return nil, fmt.Errorf("no storage found")
	}

	return &CLI{
		device:  dev,
		storage: storages[0].Sid,
	}, nil
}

func printUsage() {
	fmt.Println("Usage: mtpx-cli <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  list <remote_path>                  List files at remote path")
	fmt.Println("  download <remote> <local_dir>       Download a file into target directory")
	fmt.Println("  upload <local_file> <remote_dir>    Upload a file into remote directory")
	fmt.Println("  delete <remote_path> [...]          Delete one or more files by remote path")
	fmt.Println("  stat <remote_path>                  Check if a file exists and print its size")
	fmt.Println("  device-info                         Show basic device information")
	fmt.Println("  storage-info                        Show storage-related information")
}

// JSON output helpers
func printJSON(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

func printProgress(file string, progress float64) error {
	return printJSON(map[string]interface{}{
		"file":     file,
		"progress": progress,
	})
}

func printTransferSummary(source, target string) error {
	return printJSON(map[string]string{
		"source": source,
		"target": target,
	})
}

// Command handlers
func (c *CLI) handleList(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("list requires remote path")
	}

	_, _, _, err := mtpx.Walk(c.device, c.storage, args[0], true, true, false,
		func(objectId uint32, fi *mtpx.FileInfo, err error) error {
			if err == nil {
				printJSON(map[string]interface{}{
					"path": fi.FullPath,
					"size": fi.Size,
				})
			}
			return nil
		})
	
	if err != nil {
		return err
	}
	
	fmt.Println("MTPX_LIST_DONE")
	return nil
}

func (c *CLI) handleDownload(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("download requires remote path and local target dir")
	}

	targetDir, err := filepath.Abs(args[1])
	if err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	handler := &ProgressHandler{targetDir: targetDir}
	
	_, _, err = mtpx.DownloadFiles(c.device, c.storage, []string{args[0]}, targetDir, false,
		func(fi *mtpx.FileInfo, err error) error { return nil },
		handler.handleDownloadProgress)
	
	if err != nil {
		return err
	}
	
	fmt.Println("MTPX_DOWNLOAD_DONE")
	return nil
}

func (c *CLI) handleUpload(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("upload requires local file and remote target dir")
	}

	localFile, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("invalid local file path: %w", err)
	}

	handler := &ProgressHandler{
		sourcePath: localFile,
		targetDir:  args[1],
	}
	
	_, _, _, err = mtpx.UploadFiles(c.device, c.storage, []string{localFile}, args[1], false,
		func(fi *os.FileInfo, path string, err error) error { return nil },
		handler.handleUploadProgress)
	
	if err != nil {
		return err
	}
	
	fmt.Println("MTPX_UPLOAD_DONE")
	return nil
}

func (c *CLI) handleDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("delete requires at least one remote path")
	}

	var props []mtpx.FileProp
	for _, path := range args {
		props = append(props, mtpx.FileProp{FullPath: path})
	}

	err := mtpx.DeleteFile(c.device, c.storage, props)
	if err != nil {
		return err
	}

	fmt.Println("MTPX_DELETE_DONE")
	return nil
}

func (c *CLI) handleStat(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("stat requires a remote path")
	}

	props := []mtpx.FileProp{{FullPath: args[0]}}
	results, err := mtpx.FileExists(c.device, c.storage, props)
	if err != nil {
		return err
	}

	info := results[0]
	if info.Exists {
		fi := info.FileInfo
		fmt.Printf("STAT\t%s\t%d\t%s\n", fi.FullPath, fi.Size, humanReadableSize(fi.Size))
	} else {
		fmt.Println("NOT_FOUND")
	}
	
	fmt.Println("MTPX_STAT_DONE")
	return nil
}

func (c *CLI) handleDeviceInfo(args []string) error {
	info, err := mtpx.FetchDeviceInfo(c.device)
	if err != nil {
		return err
	}

	printJSON(info)
	fmt.Println("MTPX_DEVICE_INFO_DONE")
	return nil
}

func (c *CLI) handleStorageInfo(args []string) error {
	storages, err := mtpx.FetchStorages(c.device)
	if err != nil {
		return fmt.Errorf("failed to fetch storage info: %w", err)
	}

	printJSON(storages)
	fmt.Println("MTPX_STORAGE_INFO_DONE")
	return nil
}

// Progress handlers
func (p *ProgressHandler) handleDownloadProgress(pi *mtpx.ProgressInfo, err error) error {
	if pi.ActiveFileSize.Progress < 100.0 {
		printProgress(pi.FileInfo.Name, float64(pi.ActiveFileSize.Progress))
	} else if pi.ActiveFileSize.Progress == 100.0 && !p.printedDone {
		printProgress(pi.FileInfo.Name, 100.0)
		p.printedDone = true
		
		sourcePath := pi.FileInfo.FullPath
		targetPath := filepath.Join(p.targetDir, pi.FileInfo.Name)
		printTransferSummary(sourcePath, targetPath)
	}
	return nil
}

func (p *ProgressHandler) handleUploadProgress(pi *mtpx.ProgressInfo, err error) error {
	if pi.ActiveFileSize.Progress < 100.0 {
		printProgress(pi.FileInfo.Name, float64(pi.ActiveFileSize.Progress))
	} else if pi.ActiveFileSize.Progress == 100.0 && !p.printedDone {
		printProgress(pi.FileInfo.Name, 100.0)
		p.printedDone = true
		
		targetPath := filepath.Join(p.targetDir, pi.FileInfo.Name)
		printTransferSummary(p.sourcePath, targetPath)
	}
	return nil
}

// Utility functions
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