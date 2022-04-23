package kamstrup

type RegisterValue interface {
	Error() error
	Unit() string
	Value() float64
}

type regValue struct {
	error error
	unit  string
	value float64
}

func errorValue(error error) RegisterValue {
	return regValue{
		error: error,
	}
}

func registerValue(val float64, unit string) RegisterValue {
	return regValue{
		unit:  unit,
		value: val,
	}
}

func (v regValue) Error() error {
	return v.error
}

func (v regValue) Unit() string {
	return v.unit
}

func (v regValue) Value() float64 {
	return v.value
}
