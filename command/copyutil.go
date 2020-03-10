package command

import (
	"context"
	"fmt"

	"github.com/peak/s5cmd/objurl"
	"github.com/peak/s5cmd/storage"
)

// checkConditions checks if the job satisfies the conditions if the job has -n, -s and -u flags.
// It returns error-embedded JobResponse with status "warning" if none of the requirements are met.
// It returns nil if any warning or error is encountered during this check.
func checkConditions(
	ctx context.Context,
	src *objurl.ObjectURL,
	dst *objurl.ObjectURL,
	noClobber bool,
	ifSizeDiffer bool,
	ifSourceNewer bool,
) error {
	// if has no flag, return nil
	if !noClobber && !ifSizeDiffer && !ifSourceNewer {
		return nil
	}

	srcObj, err := getObject(ctx, src)
	if err != nil {
		return err
	}

	dstObj, err := getObject(ctx, dst)
	if err != nil {
		return err
	}

	// if destination is not exists, no conditions apply.
	if dstObj == nil {
		return nil
	}

	var stickyErr error
	if noClobber {
		stickyErr = ErrObjectExists
	}

	if ifSizeDiffer {
		if srcObj.Size == dstObj.Size {
			stickyErr = ErrObjectSizesMatch
		} else {
			stickyErr = nil
		}
	}

	if ifSourceNewer {
		srcMod, dstMod := srcObj.ModTime, dstObj.ModTime

		if !srcMod.After(*dstMod) {
			stickyErr = ErrObjectIsNewer
		} else {
			stickyErr = nil
		}
	}

	return stickyErr
}

// getObject checks if the object from given url exists. If no object is
// found, error and returning object would be nil.
func getObject(ctx context.Context, url *objurl.ObjectURL) (*storage.Object, error) {
	client, err := storage.NewClient(url)
	if err != nil {
		return nil, err
	}

	obj, err := client.Stat(ctx, url)
	if err == storage.ErrGivenObjectNotFound {
		return nil, nil
	}

	return obj, err
}

//  OK-to-have error types (warnings) that is used when the job status is warning.
var (
	ErrObjectExists     = fmt.Errorf("object already exists")
	ErrObjectIsNewer    = fmt.Errorf("object is newer or same age")
	ErrObjectSizesMatch = fmt.Errorf("object size matches")
)

func isWarning(err error) bool {
	switch err {
	case ErrObjectExists, ErrObjectIsNewer, ErrObjectSizesMatch:
		return true
	}

	return false
}
