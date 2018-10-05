package main

/*===================import=====================*/

import (
	"bufio"
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"os"
	"os/exec"
	"strings"
)

/*====================type===================*/

type selpg_Args struct {
	start_page   int
	end_page     int
	page_len     int
	page_type    string // 'l' for lines-delimited, 'f' for form-feed-delimited default is 'l'
	print_dest   string
	input_source string // 输入途径，默认为键盘输入

}

/*====================function=================*/

//main函数调用
func main() {
	sa := new(selpg_Args)

	//参数绑定变量
	flag.IntVar(&sa.start_page, "s", -1, "the start Page")
	flag.IntVar(&sa.end_page, "e", -1, "the end Page")
	flag.IntVar(&sa.page_len, "l", 72, "the length of the page")
	flag.StringVar(&sa.print_dest, "d", "", "the destiny of printing") //默认值缺省

	/*检查命令中是否含有-f
	如果有，则selpg在输入中寻找换页符，并将其作为页定界符
	若没有，则按照输入的-l的长度作为页的长度
	*/
	exist_f := flag.Bool("f", false, "")
	flag.Parse()

	//如果命令中使用了-f
	if *exist_f {
		sa.page_type = "f"
		sa.page_len = -1
	} else { //如果没有，则按照l规定的页长
		sa.page_type = "l"
	}

	//初始化输入来源，默认是空
	sa.input_source = ""
	//如果使用了文件名，则非标志参数的数量为1，参数为文件名
	if flag.NArg() == 1 {
		sa.input_source = flag.Arg(0)
	}

	//检查参数的合法性并执行命令
	checkArgs(*sa, flag.NArg()) //flag.Narg()返回标志处理后剩余的参数数量。
	runSelpg(*sa)
}

//检查输入参数的合法性
func checkArgs(sa selpg_Args, NArgs int) {
	//检查NArg,是标志处理后剩余的参数数量。
	Used_ok := NArgs == 1 || NArgs == 0
	//检查输入的startPage和endPage是否符合逻辑
	logic_ok := sa.start_page <= sa.end_page && sa.start_page >= 1
	//检查-f和-l的互斥性
	lf_ok := sa.page_type == "f" && sa.page_len != -1 //此时等于同时使用了-f和-l，因此是错误的
	//
	if !Used_ok || !logic_ok || lf_ok {
		usage() //提示信息
		os.Exit(1)
	}
}

//执行指令
func runSelpg(sa selpg_Args) {
	//初始化
	fin := os.Stdin
	fout := os.Stdout
	cur_page := 1 //当前页
	cur_line := 0 //当前行
	var inpipe io.WriteCloser
	var err error

	//判断是键盘读入or文件读入
	if sa.input_source != "" { //非默认->文件
		fin, err = os.Open(sa.input_source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: input file \"%s\" does not exist!\n", sa.input_source)
			//fmt.Println(err)
			usage()
			os.Exit(1)
		}
		defer fin.Close()
	}

	//确定输出到文件或者输出到屏幕
	//通过用管道接通grep模拟打印机测试，结果输出到屏幕
	if sa.print_dest != "" {
		cmd := exec.Command("grep", "-nf", "keyword")
		inpipe, err = cmd.StdinPipe()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer inpipe.Close() //最后执行
		cmd.Stdout = fout
		cmd.Start()
	}

	//判断分页方式是-l还是-f
	//按照-l来分页
	if sa.page_type == "l" {
		line := bufio.NewScanner(fin)
		//按行来读
		for line.Scan() {
			//还没输出完
			if cur_page >= sa.start_page && cur_page <= sa.end_page {
				//输出到cmd窗口
				fout.Write([]byte(line.Text() + "\n"))
				if sa.print_dest != "" {
					//输出到文件
					inpipe.Write([]byte(line.Text() + "\n"))
				}
			}
			cur_line++
			if cur_line == sa.page_len {
				cur_page++
				cur_line = 0
			}
		}
	} else { //按照-f来分页
		rd := bufio.NewReader(fin)
		for {
			//按页读
			page, ferr := rd.ReadString('\f')
			if ferr != nil || ferr == io.EOF {
				if ferr == io.EOF {
					if cur_page >= sa.start_page && cur_page <= sa.end_page {
						fmt.Fprintf(fout, "%s", page)
					}
				}
				break
			}
			page = strings.Replace(page, "\f", "", -1)
			cur_page++
			if cur_page >= sa.start_page && cur_page <= sa.end_page {
				fmt.Fprintf(fout, "%s", page)
			}
		}
	}
	//当输出完成后，比较输出的页数与期望输出的数量
	if cur_page < sa.end_page {
		fmt.Fprintf(os.Stderr, "./selpg: end_page (%d) greater than total pages (%d), less output than expected\n", sa.end_page, cur_page)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUSAGE: ./selpg [--s start_page] [--e end_page] [--l lines_per_page | --f ] [ --d dest ] [ in_filename ]\n")
}
