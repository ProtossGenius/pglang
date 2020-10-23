package snreader

import (
	"fmt"
)

func iserr(err error) bool {
	return err != nil
}

func noerr(err error) bool {
	return err == nil
}

const (
	ErrNoMatchStateNode     = "ErrNoMatchStateNode: err list: [%s]"                               //没有满足的
	ErrTooMuchStateNodeLive = "ErrTooMuchStateNodeLive, input[%v] live nodes [%s], end nodes[%s]" //太多满足条件的
)

func OnErr(reader StateNodeReader, input InputItf, reason string) error {
	return fmt.Errorf("Error in [%s], input is [%v] reason is: %s", reader.Name(), input, reason)
}
