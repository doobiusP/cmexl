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

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type templateData struct {
	UseVcpkg bool

	Name  string
	SName string
	UName string
	LName string

	Version string
}

func dirExists(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func fileExists(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

func defaultTemplateCopy(templateDir string, destRoot string, bootData *templateData) error {
	warningColor := color.New(color.FgYellow)

	processTemplate := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(templateDir, path)
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

			exists, err := dirExists(destPath)
			if err != nil {
				return err
			}
			if exists && destPath != destRoot {
				warningColor.Printf("%s is attempting to be created again\n", destPath)
			}
			err = os.MkdirAll(destPath, 0755)
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
			exists, err := fileExists(destPath)
			if err != nil {
				return err
			}
			if exists {
				warningColor.Printf("%s is attempting to be created again\n", destPath)
			}
			err = os.WriteFile(destPath, finalBuf.Bytes(), 0644)
			if err != nil {
				return err
			}
		}
		fmt.Printf("Finished creating %s\n", destPath)
		return nil
	}

	err := filepath.WalkDir(templateDir, processTemplate)
	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	return nil
}

func copyTemplates(templatesToCopy []string, bootData *templateData) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	destRoot := filepath.Join(cwd, bootData.Name)
	fmt.Printf("Creating destination root at %s\n", destRoot)
	err = os.MkdirAll(destRoot, 0755)
	if err != nil {
		return err
	}
	for _, templateDir := range templatesToCopy {
		fmt.Printf("Now copying from %s...\n", templateDir)
		err := defaultTemplateCopy(templateDir, destRoot, bootData)
		if err != nil {
			return fmt.Errorf("template %s had error %w", templateDir, err)
		}
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
	if strings.Contains(name, " ") {
		return errors.New("name cannot contain spaces")
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
	exists, err := dirExists(templateDir)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("template provided not found in cmexl/templates")
	}

	templatesToCopy := []string{templateDir}

	if !noVcpkg {
		vcpkgDir := filepath.Join(execDir, "templates", "vcpkg_common")
		exists, err = dirExists(vcpkgDir)

		if err != nil {
			return err
		}
		if !exists {
			return errors.New("vcpkg_common provided not found in cmexl/templates")
		}
		templatesToCopy = append(templatesToCopy, vcpkgDir)
	}

	sname, snameErr := cmd.PersistentFlags().GetString("short-name")
	if snameErr != nil {
		return snameErr
	}

	if len(name) <= 0 {
		return errors.New("name must be present and non-empty")
	}
	if len(sname) <= 0 {
		sname = name
	}
	if strings.Contains(sname, " ") {
		return errors.New("short-name cannot contain spaces")
	}

	bootData := templateData{
		UseVcpkg: !noVcpkg,
		Name:     name,
		SName:    sname,
		UName:    strings.ToUpper(sname),
		LName:    strings.ToLower(sname),
		Version:  "0.1.0.0",
	}

	err = copyTemplates(templatesToCopy, &bootData)

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

	// TODO: omit version handling flag maybe?
	// TODO: specify, add, delete configs.
}
