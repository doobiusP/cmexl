package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

type initData struct {
	UseVcpkg    bool
	TemplateDir string

	Name  string
	SName string
	UName string
	LName string

	Version string
}

func bootstrapTemplate(bootData *initData) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	destRoot := filepath.Join(cwd, bootData.Name)
	os.MkdirAll(destRoot, 0755)

	processTemplate := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(bootData.TemplateDir, path)
		if err != nil {
			return err
		}
		var destPath string
		if d.IsDir() {
			var buf bytes.Buffer
			var pathTmpl = template.Must(template.New("").Parse(rel))
			if err := pathTmpl.Execute(&buf, bootData); err != nil {
				return err
			}
			newRel := buf.String()
			destPath = filepath.Join(destRoot, newRel)
			err := os.MkdirAll(destPath, 0755)
			if err != nil {
				return err
			}
		} else {
			srcBytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var contentTmpl = template.Must(template.New("").Parse(string(srcBytes)))
			var finalBuf bytes.Buffer
			if err := contentTmpl.Execute(&finalBuf, bootData); err != nil {
				return err
			}

			var filenameTmpl = template.Must(template.New("").Parse(rel))
			var fileBuf bytes.Buffer
			if err := filenameTmpl.Execute(&fileBuf, bootData); err != nil {
				return err
			}
			newRel := fileBuf.String()
			newRel = strings.TrimSuffix(newRel, ".go.tmpl")
			destPath = filepath.Join(destRoot, newRel)

			err = os.WriteFile(destPath, finalBuf.Bytes(), 0644)
			if err != nil {
				return err
			}
		}
		fmt.Println(destPath)
		return nil
	}

	err = filepath.WalkDir(bootData.TemplateDir, processTemplate)
	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	return nil
}

func initE(cmd *cobra.Command, args []string) error {
	noVcpkg, noVcpkgErr := cmd.PersistentFlags().GetBool("no-vcpkg")
	if noVcpkgErr != nil {
		return noVcpkgErr
	}
	name, nameErr := cmd.PersistentFlags().GetString("name")
	if nameErr != nil {
		return nameErr
	}
	template, templateErr := cmd.PersistentFlags().GetString("template")
	if templateErr != nil {
		return templateErr
	}
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(execPath)
	templateDir := filepath.Join(execDir, "templates", template)
	if info, err := os.Stat(templateDir); err != nil || !info.IsDir() {
		return errors.New("template provided not found in cmexl/templates")
	}

	sname, snameErr := cmd.PersistentFlags().GetString("short-name")
	if snameErr != nil {
		return snameErr
	}
	version, versionErr := cmd.PersistentFlags().GetString("version")
	if versionErr != nil {
		return versionErr
	}

	if len(name) <= 0 {
		return errors.New("name must be present and non-empty")
	}
	if len(sname) <= 0 {
		sname = name
	}

	// TODO: ensure appropriate version format
	if len(version) <= 0 {
		version = "0.1.0.0"
	}

	bootData := initData{
		UseVcpkg:    !noVcpkg,
		TemplateDir: templateDir,
		Name:        name,
		SName:       sname,
		UName:       strings.ToUpper(sname),
		LName:       strings.ToLower(sname),
		Version:     version,
	}

	err = bootstrapTemplate(&bootData)
	if err != nil {
		return err
	}
	return err
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstrap a new cmexl project from a template",
	RunE:  initE,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.PersistentFlags().String("name", "", "Formal name of project. Required")
	initCmd.MarkPersistentFlagRequired("name")

	initCmd.PersistentFlags().String("template", "", "Template to bootstrap project with. Required")
	initCmd.MarkPersistentFlagRequired("template")

	initCmd.PersistentFlags().Bool("no-vcpkg", false, "Omit vcpkg details. Default false")
	initCmd.PersistentFlags().String("short-name", "", "Short name of project used in generated files. Default <name>")
	initCmd.PersistentFlags().String("version", "0.1.0.0", "Initial version of project. Follows MAJOR.MINOR.PATCH.TWEAK. Default 0.1.0.0")

	// TODO: omit version handling flag maybe?
	// TODO: specify, add, delete configs.
}
