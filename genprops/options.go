package genprops

import "gorm.io/gorm/schema"

type Option func(*config)

type config struct {
	outputDir      string
	outputFile     string
	packageName    string
	namingStrategy schema.NamingStrategy
	dryRun         bool
	logger         Logger
}

func defaultConfig() config {
	return config{
		outputDir:  "./",
		outputFile: "props_gen.go",
		namingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		logger: NewDefaultLogger(),
	}
}

// WithOutputDir sets the output directory.
//
// Example:
//
//	g := genprops.New(genprops.WithOutputDir("./model"))
//	_ = g
func WithOutputDir(dir string) Option {
	return func(c *config) {
		c.outputDir = dir
	}
}

// WithOutputFile sets the output file name.
//
// Example:
//
//	g := genprops.New(genprops.WithOutputFile("props_gen.go"))
//	_ = g
func WithOutputFile(name string) Option {
	return func(c *config) {
		c.outputFile = name
	}
}

// WithPackageName sets the output package name.
//
// Example:
//
//	g := genprops.New(genprops.WithPackageName("model"))
//	_ = g
func WithPackageName(name string) Option {
	return func(c *config) {
		c.packageName = name
	}
}

// WithNamingStrategy sets the GORM naming strategy used when parsing schemas.
//
// Example:
//
//	ns := schema.NamingStrategy{SingularTable: true}
//	g := genprops.New(genprops.WithNamingStrategy(ns))
//	_ = g
func WithNamingStrategy(ns schema.NamingStrategy) Option {
	return func(c *config) {
		c.namingStrategy = ns
	}
}

// WithDryRun enables/disables dry-run mode.
//
// Example:
//
//	err := genprops.New(genprops.WithDryRun(true)).Generate(&model.User{})
//	_ = err
func WithDryRun(d bool) Option {
	return func(c *config) {
		c.dryRun = d
	}
}

// WithLogger sets the logger. Passing nil uses a no-op logger.
//
// Example:
//
//	g := genprops.New(genprops.WithLogger(genprops.NewDefaultLogger()))
//	_ = g
func WithLogger(l Logger) Option {
	return func(c *config) {
		if l == nil {
			c.logger = NopLogger()
		} else {
			c.logger = l
		}
	}
}
