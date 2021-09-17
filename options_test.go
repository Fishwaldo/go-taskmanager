package taskmanager

import (
	"testing"

)

func TestOptionsLogger(t *testing.T) {
	options := defaultTaskOptions()
	log := DefaultLogger()
	logop := WithLogger(log)
	logop.apply(options)
	switch options.logger.(type) {
	case *StdLogger:
	default:
		t.Errorf("WithLogger Options Apply Failed")
	}
}

type testemw struct {
}
func (mw *testemw) PreHandler(s *Task) (MWResult,error) {
	return MWResult{}, nil
}
func (mw *testemw) PostHandler(s *Task, err error) (MWResult) {
	return MWResult{}
}
func (mw *testemw) Reset(s *Task) {

}
func (mw *testemw) Initilize(s *Task) {

}
func TestOptionsExecutionMW (t *testing.T) {
	options := defaultTaskOptions()
	mw := &testemw{}
	mwop := WithExecutationMiddleWare(mw)
	mwop.apply(options)
	switch options.executationmiddlewares[0].(type) {
	case *testemw:
	default:
		t.Errorf("WithExecutionMiddleware isn't testmw")
	}
}

type testrmw struct {
	
	
}

func (mw *testrmw) Handler(s *Task, prerun bool, e error) (retry RetryResult, err error) {
	return RetryResult{}, nil
} 

func (mw *testrmw) 	Reset(s *Task) (ok bool) {
	return false
}

func (mw *testrmw) Initilize(s *Task) {
}

func TestOptionsRetryMW (t *testing.T) {
	options := defaultTaskOptions()
	mw := &testrmw{}
	mwop := WithRetryMiddleWare(mw)
	mwop.apply(options)
	switch options.retryMiddlewares[0].(type) {
	case *testrmw:
	default:
		t.Errorf("WithRetryMiddleWare isn't testmw")
	}
}