package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	ct "github.com/daviddengcn/go-colortext"
)

var (
	RootDirName = "C:\\Users\\jumang4423\\syscall_go_f_crwd\\entity_test_dir"
	difference  = []DiffFileObj{} // 違いを保持するグローバル変数
)

// ファイルのオブジェクト
type FileObj struct {
	Name              string
	ModifiedTimeStamp *syscall.Filetime
	AccessedTimeStamp *syscall.Filetime
}

// フォルダのオブジェクト
type DirObj struct {
	Files   []FileObj
	Bk      *[]DirObj
	DirPath string // パスの保時
}

// 比較して、違いを保持する
type DiffFileObj struct {
	FilePath string
	isCreate bool
	isWrite  bool
	isRead   bool
	isDir    bool
}

// ディレクトリツリーを作成する
func BuildDirTree(dir string, bk *DirObj) {
	var files_obj []FileObj
	var dirs_obj []DirObj

	// ++ 表示
	dir_array := strings.Split(dir, "\\")
	for i := 0; i < len(dir_array); i++ {
		print("    ")
	}
	print("@", dir_array[len(dir_array)-1], "\n")

	// ディレクトリに何があるか見る
	f_or_d, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, file := range f_or_d {
		// ファイルの情報を見る
		fileInfo, err := os.Stat(dir + "\\" + file.Name())
		// ディレクトリならダメ
		if err != nil || fileInfo.IsDir() {
			continue
		}

		// ファイルの情報を取得する
		fileObj := FileObj{
			Name:              fileInfo.Name(),
			ModifiedTimeStamp: &fileInfo.Sys().(*syscall.Win32FileAttributeData).LastWriteTime,
			AccessedTimeStamp: &fileInfo.Sys().(*syscall.Win32FileAttributeData).LastAccessTime,
		}

		// ファイルを追加する
		files_obj = append(files_obj, fileObj)

		// 表示
		for i := 0; i < len(dir_array); i++ {
			print("    ")
		}
		print("L", fileObj.Name, "\n")
	}

	for _, child_dir := range f_or_d {
		// ディレクトリじゃなかったら
		if !child_dir.IsDir() {
			continue
		}

		// ディレクトリの情報を取得する（佐伯）
		var dirs_obj_recursive DirObj
		BuildDirTree(dir+"\\"+child_dir.Name(), &dirs_obj_recursive)

		// ディレクトリ達を追加する
		dirs_obj = append(dirs_obj, dirs_obj_recursive)
	}

	// ディレクトリを追加する
	*bk = DirObj{Files: files_obj, Bk: &dirs_obj, DirPath: dir}
}

// 昔のファイルに今のファイルがあるか確認して、そのファイルのindexを返す
func FindFile(dir string, files *[]FileObj) (int, bool) {
	for index, file := range *files {
		if file.Name == dir {
			return index, true
		}
	}

	return -1, false
}

// 昔のフォルダーに今のフォルダーがあるか確認して、そのフォルダーのindexを返す
func FindDir(dir string, dirs *[]DirObj) (int, bool) {
	for index, dir_str := range *dirs {
		if dir_str.DirPath == dir {
			return index, true
		}
	}

	return -1, false
}

// 昔のバックアップと今のツリーを比較する
func CompareOldBkAndCurrentBk(dir string, oldBk *DirObj) {
	// 表示
	splited_dir := strings.Split(dir, "\\")
	for i := 0; i < len(splited_dir); i++ {
		print("    ")
	}
	print("@", splited_dir[len(splited_dir)-1], "\n")

	// ディレクトリに何があるか見る
	f_or_d, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	// ファイルの情報を見る
	for _, file := range f_or_d {

		// ディレクトリならダメ、ファイルチェック
		fileInfo, err := os.Stat(dir + "\\" + file.Name())
		if err != nil || fileInfo.IsDir() {
			continue
		}

		// ファイルの情報を取得する
		index, found := FindFile(file.Name(), &oldBk.Files)
		if !found {
			println("file created")
			os.Exit(1)
		}
		// ファイルの情報を取得する
		fileObj := FileObj{
			Name:              fileInfo.Name(),
			ModifiedTimeStamp: &fileInfo.Sys().(*syscall.Win32FileAttributeData).LastWriteTime,
			AccessedTimeStamp: &fileInfo.Sys().(*syscall.Win32FileAttributeData).LastAccessTime,
		}

		// ファイルの情報から比較する
		isWrite := false
		isRead := false

		if fileObj.AccessedTimeStamp != oldBk.Files[index].AccessedTimeStamp {
			isRead = true

		}
		if fileObj.ModifiedTimeStamp != oldBk.Files[index].ModifiedTimeStamp {
			isWrite = true
		}

		// もしファイルが更新されていたら、differenceに追加する
		ct.Foreground(ct.White, false)
		if isWrite || isRead {
			//差分が検知されたところは緑色がいいな
			ct.Foreground(ct.Green, false)
			for i := 0; i < len(splited_dir); i++ {
				print("    ")
			}
			print("L", fileInfo.Name())
			if isWrite {
				print("(W)")
			}
			if isRead {
				print("(R)")
			}

			print("\n")
			difference = append(difference, DiffFileObj{FilePath: dir + "\\" + fileObj.Name, isWrite: isWrite, isRead: isRead, isCreate: false, isDir: false})

			// 昔のファイルの情報に対して今の情報に上書きする
			oldBk.Files[index] = fileObj
		}
		ct.ResetColor()
	}

	// ディレクトリに何があるか見る
	for _, child_dir := range f_or_d {

		// ディレクトリじゃなかったら
		if !child_dir.IsDir() {
			continue
		}

		// ディレクトリのINDEXを取得する
		index, found := FindDir(dir+"\\"+child_dir.Name(), oldBk.Bk)

		if !found {
			// 新たに作られたディレクトリらしい
			println("dir created")
			os.Exit(1)
		}

		// ディレクトリの情報を取得する（佐伯）
		CompareOldBkAndCurrentBk(dir+"\\"+child_dir.Name(), &(*oldBk.Bk)[index])
	}
}

//  ターミナルクリあ
func clear_terminal() {
	cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	// 検索開始
	var rootDir DirObj
	clear_terminal()
	BuildDirTree(RootDirName, &rootDir)

	for {
		// １秒待って
		time.Sleep(time.Second * 1)
		// 差分検知
		clear_terminal()
		CompareOldBkAndCurrentBk(RootDirName, &rootDir)
	}
}
