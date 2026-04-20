package genprops

import "gorm.io/gorm/schema"

type Option func(*config)

type config struct {
	outputDir      string
	outputFile     string
	packageName    string
	namingStrategy schema.NamingStrategy
	dryRun         bool
	logger         func(format string, a ...any)
}

func defaultConfig() config {
	return config{
		outputDir:  "./",
		outputFile: "props_gen.go",
		namingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}
}

// WithOutputDir 设置输出目录
func WithOutputDir(dir string) Option {
	return func(c *config) {
		c.outputDir = dir
	}
}

// WithOutputFile 设置输出文件名
func WithOutputFile(name string) Option {
	return func(c *config) {
		c.outputFile = name
	}
}

// WithPackageName 设置输出包名
func WithPackageName(name string) Option {
	return func(c *config) {
		c.packageName = name
	}
}

// WithNamingStrategy 设置命名策略
func WithNamingStrategy(ns schema.NamingStrategy) Option {
	return func(c *config) {
		c.namingStrategy = ns
	}
}

// WithDryRun 设置为 dry run 模式
func WithDryRun(d bool) Option {
	return func(c *config) {
		c.dryRun = d
	}
}

// WithLogger 设置日志记录器
func WithLogger(fn func(format string, a ...any)) Option {
	return func(c *config) {
		c.logger = fn
	}
}
