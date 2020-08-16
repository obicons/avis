package platforms

import "context"

type System interface {
	Start() error
	Stop(ctx context.Context) error
}
