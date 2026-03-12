package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
	"gorm.io/gorm/schema"
)

var (
	typeNames = flag.String("type", "", "逗号分隔的类型名称列表；必须设置")
	output    = flag.String("output", "", "输出文件名；默认 srcdir/<type>_gen.go")
)

var namingStrategy = schema.NamingStrategy{
	SingularTable: true,
}

func main() {
	// 1. 解析参数
	flag.Parse()
	if len(*typeNames) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	goFile, dir, outName := parseEnvAndArgs()

	// 2. 加载包
	pkgs := loadPackage(goFile, dir)

	// 3. 查找目标包
	pkg := findTargetPackage(pkgs)

	// 4. 初始化生成器
	g := Generator{
		pkgName:    pkg.Name,
		data:       make(map[string][]FieldInfo),
		outputPath: outName,
	}

	// 5. 处理类型
	g.processTypes(pkg)

	// 6. 生成代码
	g.generate()
}

func parseEnvAndArgs() (string, string, string) {
	goFile := os.Getenv("GOFILE")
	var dir string
	var err error

	if goFile == "" {
		// 降级策略：如果不是通过 go:generate 运行，则解析当前目录
		dir, err = os.Getwd()
		if err != nil {
			log.Fatalf("无法获取当前目录: %v", err)
		}
		log.Println("未检测到 GOFILE 环境变量，将解析当前目录下的包")
	} else {
		dir = filepath.Dir(goFile)
	}

	// 确定输出文件名
	outName := *output
	if outName == "" {
		if goFile != "" {
			baseName := strings.TrimSuffix(goFile, ".go")
			outName = baseName + "_gen.go"
		} else {
			// 如果是目录模式，默认使用第一个类型名的小写
			typesList := strings.Split(*typeNames, ",")
			outName = strings.ToLower(typesList[0]) + "_gen.go"
		}
	}
	return goFile, dir, outName
}

func loadPackage(goFile, dir string) []*packages.Package {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedImports | packages.NeedDeps,
		Dir:  dir,
	}

	var pattern string
	if goFile != "" {
		absGoFile, err := filepath.Abs(goFile)
		if err != nil {
			log.Fatalf("无法获取 GOFILE 的绝对路径: %v", err)
		}
		pattern = "file=" + absGoFile
	} else {
		pattern = "."
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		log.Fatalf("加载包失败: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}
	return pkgs
}

func findTargetPackage(pkgs []*packages.Package) *packages.Package {
	typesList := strings.Split(*typeNames, ",")
	var pkg *packages.Package
	var foundPkg bool

	for _, p := range pkgs {
		for _, typeName := range typesList {
			if obj := p.Types.Scope().Lookup(typeName); obj != nil {
				pkg = p
				foundPkg = true
				break
			}
		}
		if foundPkg {
			break
		}
	}

	if !foundPkg {
		if len(pkgs) > 0 {
			pkg = pkgs[0]
		} else {
			log.Fatal("未找到任何包")
		}
	}
	return pkg
}
