package transport

import "context"

type Mock struct {
	WriteFunc func(ctx context.Context, data []byte) error
	ReadFunc  func(ctx context.Context) ([]byte, error)
	CloseFunc func() error
}

func (m *Mock) Write(ctx context.Context, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.WriteFunc != nil {
		return m.WriteFunc(ctx, data)
	}
	return nil
}

func (m *Mock) Read(ctx context.Context) ([]byte, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(ctx)
	}
	return nil, nil
}

func (m *Mock) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

type SeqMock struct {
	Reads   []func() ([]byte, error)
	ReadIdx int

	Writes    []func() error
	WriteIdx  int
	WritesLog [][]byte

	CloseFunc func() error
}

func (m *SeqMock) Read(ctx context.Context) ([]byte, error) {
	if m.ReadIdx >= len(m.Reads) {
		return nil, nil
	}
	resp, err := m.Reads[m.ReadIdx]()
	m.ReadIdx++
	return resp, err
}

func (m *SeqMock) Write(ctx context.Context, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.WritesLog = append(m.WritesLog, data)
	if m.WriteIdx >= len(m.Writes) {
		return nil
	}
	err := m.Writes[m.WriteIdx]()
	m.WriteIdx++
	return err
}

func (m *SeqMock) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
