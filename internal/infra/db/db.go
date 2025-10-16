package db

type scannable interface {
	Scan(dest ...any) error
}
