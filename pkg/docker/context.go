package docker

import (
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
	"path/filepath"
)

//Sets a docker context for a specific host.
func (d *dockerImpl) ContextSet(
	name string,
	dockerHost string,
) error {
	dockerConfigPath, err := d.dockerConfigFolderPath()

	if err != nil {
		return errors.Wrap(err, "failed to get docker config path")
	}

	contextFolderName, err := d.contextFolderHashedName(name)

	if err != nil {
		return errors.Wrap(err, "failed to generate context folder name")
	}

	contextFolderPath := filepath.Join(
		*dockerConfigPath, "contexts", "meta", *contextFolderName,
	)
	metaFilePath := filepath.Join(contextFolderPath, "meta.json")

	if contextFolderExists, err := d.os.PathExists(contextFolderPath); err != nil {
		return errors.Wrapf(err, "failed to check %s path existence", contextFolderPath)
	} else if !contextFolderExists {
		if err := d.os.MkdirAll(contextFolderPath, 0700); err != nil {
			return errors.Wrap(err, "failed to create context folder")
		}
	}

	context := d.createContextMeta(name, dockerHost)
	serialized, err := context.JSON()

	if err != nil {
		return err
	}

	err = d.io.WriteFile(metaFilePath, serialized, 0644)

	if err != nil {
		return errors.Wrap(err, "failed to write meta file")
	}

	return nil
}

//Gets the hashed value for a given context name.
func (d *dockerImpl) contextFolderHashedName(name string) (*string, error) {
	h := sha256.New()

	if _, err := h.Write([]byte(name)); err != nil {
		return nil, err
	}

	hashedName := fmt.Sprintf("%x", h.Sum(nil))
	return &hashedName, nil
}
