package skills

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func removeIfExists(path string) error {
	if _, err := os.Lstat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return os.RemoveAll(path)
}

func copyFile(src, dst string, mode fs.FileMode) error {
	if err := ensureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}

	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}

func copyDir(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		dst := filepath.Join(dstDir, rel)

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}

		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}

			if err := ensureDir(filepath.Dir(dst)); err != nil {
				return err
			}

			return os.Symlink(link, dst)
		}

		return copyFile(path, dst, info.Mode())
	})
}

func installFromDirToTargets(installName string, srcDir string, targetDirs []string, mode InstallMode) error {
	if len(targetDirs) == 0 {
		return ErrNoTargetDirs
	}

	primarySkillsDir := targetDirs[0]

	primarySkillDir := filepath.Join(primarySkillsDir, installName)
	if err := ensureDir(primarySkillsDir); err != nil {
		return err
	}

	if err := removeIfExists(primarySkillDir); err != nil {
		return err
	}

	if err := copyDir(srcDir, primarySkillDir); err != nil {
		return err
	}

	for _, skillsDir := range targetDirs[1:] {
		dst := filepath.Join(skillsDir, installName)
		if filepath.Clean(skillsDir) == filepath.Clean(primarySkillsDir) {
			continue
		}

		if err := ensureDir(skillsDir); err != nil {
			return err
		}

		if err := removeIfExists(dst); err != nil {
			return err
		}

		switch mode {
		case InstallModeSymlink:
			if err := os.Symlink(primarySkillDir, dst); err != nil {
				if err := copyDir(primarySkillDir, dst); err != nil {
					return err
				}
			}
		case InstallModeCopy:
			if err := copyDir(primarySkillDir, dst); err != nil {
				return err
			}
		default:
			return ErrUnknownInstallMode
		}
	}

	return nil
}
