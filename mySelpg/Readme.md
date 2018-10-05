# Selpg #

----------

## 设计说明 ##

###【程序简介】 
Selpg从标准输入或从作为命令行参数给出的文件名读取文本输入。它允许用户指定来自该输入并随后将被输出的页面范围,然后输出到标准输出或是文件中。

详细介绍参考[https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html](https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html "Here")

###【程序设计】
程序的功能分为以下几个部分：


- 读取一条命令行输入的指令
- 解析命令，分析其中的参数
- 实现命令请求的操作
- 输入命令有误时进行提示


####一、【读取输入指令并解析】

输入指令的格式为 
**`    selpg [--s start_page] [--e end_page] [--l lines_per_page | --f ] [ --d dest ] [ input_source ]`**

其中各个参数的含义如下:


- --s    从输入的第几页开始读起，默认值为-1
- --e	 读到输入的第几页，默认值为-1
- --l	 当换页模式为【l】时，每一页的行数，默认值为72
- --f    将换页模式设定为【f】，即遇到分页符时自动换页
- （注：--l与--f是互斥的两个属性）
- --d    输出的地址
- 【input_source】没有flag值，指输入文件的文件名，默认为空，即标准输入

使用一个结构体存储指令，成员包括以上的参数
    
    
    type selpg_Args struct {
    	start_page   int
    	end_page int
    	page_len int
    	page_typestring // 'l' for lines-delimited, 'f' for form-feed-delimited default is 'l'
    	print_dest   string
    	input_source string // 输入途径，默认为键盘输入
    
    }
    
使用【flag.XXXVar】依次绑定指令中的各个标识值与对应的变量

	flag.IntVar(&sa.start_page, "s", -1, "the start Page")
	flag.IntVar(&sa.end_page, "e", -1, "the end Page")
	flag.IntVar(&sa.page_len, "l", 72, "the length of the page")
	flag.StringVar(&sa.print_dest, "d", "", "the destiny of printing") //默认值缺省

	/*检查命令中是否含有-f
	如果有，则selpg在输入中寻找换页符，并将其作为页定界符
	若没有，则按照输入的-l的长度作为页的长度
	*/
	exist_f := flag.Bool("f", false, "")*


####二、【实现命令请求的操作】

在读取完参数后，通过参数的值来实现对应的操作。

先初始化变量，通过给定的参数决定是否调用os库中的标准输入和输出
	
    fin := os.Stdin
    fout := os.Stdout
    cur_page := 1 //当前页
    cur_line := 0 //当前行
    var inpipe io.WriteCloser
    var err error

判断输入方式

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

判断输出方式

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

判断分页方式

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


####三、【错误处理】

对于非标准格式的输入进行了提示

（使用pflag代替goflag，则输入参数时由【-s】变为【--s】）

![](https://img-blog.csdn.net/20181004231527291?watermark/2/text/aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L1llem8xMw==/font/5a6L5L2T/fontsize/400/fill/I0JBQkFCMA==/dissolve/70)

还有打开文件失败、实际输出页数小于-l的参数等错误都会有对应的提示



----------

###【测试结果】
一共创建了4个文档，分别是【input1】、【input2】、【output1】、【error_file】


- 【input1】是一个20行的输入文档，每行的内容为【line+数字】
- 【input2】内容与【input1】相仿，但每隔4行的结尾插入一个换页符【/f】
- 【output1】是一个空文档，用作输出文档
- 【error_file】用于记录错误信息，可将错误信息输出到该文档

测试文档存放在名为【test_file】的文件夹中

测试结果以截图的形式存放在名为【test_result】的文件夹中
