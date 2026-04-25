package colgen

import "gorm.io/gorm/schema"

type Option func(*config)

type config struct {
	outputDir      string
	packageName    string
	namingStrategy schema.NamingStrategy
	dryRun         bool
	logger         Logger
}

func defaultConfig() config {
	return config{
		outputDir: "columns",
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
//	g := colgen.New(colgen.WithOutputDir("./model"))
//	_ = g
func WithOutputDir(dir string) Option {
	return func(c *config) {
		c.outputDir = dir
	}
}

// WithPackageName sets the output package name.
//
// Example:
//
//	g := colgen.New(colgen.WithPackageName("model"))
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
//	g := colgen.New(colgen.WithNamingStrategy(ns))
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
//	err := colgen.New(colgen.WithDryRun(true)).Generate(&model.User{})
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
//	g := colgen.New(colgen.WithLogger(colgen.NewDefaultLogger()))
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
