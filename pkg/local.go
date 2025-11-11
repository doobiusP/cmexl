package cmexl_utils

import "os"

func CreateCmexlStore() error {
	err := os.MkdirAll(".cmexl", 0755)
	return err
}
