package transform

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/multierr"
)

var (
	BrandValue    = "Toyota"
	YearValue     = 2020
	UsedValue     = true
	PriceValue    = 100000.50
	DocumentValue = "123"
)

type Car struct {
	Brand    string
	Year     int
	Used     bool
	Price    float64
	Document string
}

type unmarshalTestSuite struct {
	suite.Suite
	data map[string]interface{}
	car  Car
	cars []Car
}

func (suite *unmarshalTestSuite) SetupTest() {
	suite.data = map[string]interface{}{
		"Brand":    BrandValue,
		"Year":     YearValue,
		"Used":     UsedValue,
		"Price":    PriceValue,
		"Document": DocumentValue,
	}
}

func (suite *unmarshalTestSuite) CarAssertions() {
	assert.Equal(suite.T(), BrandValue, suite.car.Brand)
	assert.Equal(suite.T(), YearValue, suite.car.Year)
	assert.Equal(suite.T(), UsedValue, suite.car.Used)
	assert.Equal(suite.T(), PriceValue, suite.car.Price)
	assert.Equal(suite.T(), DocumentValue, suite.car.Document)
}

func (suite *unmarshalTestSuite) CarsAssertions() {
	for i := 0; i < len(suite.cars); i++ {
		assert.Equal(suite.T(), BrandValue, suite.cars[i].Brand)
		assert.Equal(suite.T(), YearValue, suite.cars[i].Year)
		assert.Equal(suite.T(), UsedValue, suite.cars[i].Used)
		assert.Equal(suite.T(), PriceValue, suite.cars[i].Price)
		assert.Equal(suite.T(), DocumentValue, suite.cars[i].Document)
	}
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseMap() {
	err := Unmarshal(suite.data, &suite.car)
	suite.Assert().Nil(err)
	suite.CarAssertions()
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseInterface() {
	var data interface{} = suite.data
	err := Unmarshal(data, &suite.car)
	suite.Assert().Nil(err)
	suite.CarAssertions()
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseByte() {
	data := []byte(`{"Brand":"Toyota","Year":2020,"Used":true,"Price":100000.50,"Document":"123"}`)
	err := Unmarshal(data, &suite.car)
	suite.Assert().Nil(err)
	suite.CarAssertions()
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseString() {
	data := `{"Brand":"Toyota","Year":2020,"Used":true,"Price":100000.50,"Document":"123"}`
	err := Unmarshal(data, &suite.car)
	suite.Assert().Nil(err)
	suite.CarAssertions()
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseSlice() {
	data := []interface{}{suite.data, suite.data}
	err := Unmarshal(data, &suite.cars)
	suite.Assert().Nil(err)
	suite.CarsAssertions()
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseSliceOfMaps() {
	data := []map[string]interface{}{suite.data, suite.data}
	err := Unmarshal(data, &suite.cars)
	suite.Assert().Nil(err)
	suite.CarsAssertions()
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseMap_DifferentType() {
	suite.data["Year"] = "2020"       // proposal: auto convert string to int
	suite.data["Price"] = "100000.50" // proposal: auto convert string to float64
	suite.data["Document"] = 123      // proposal: auto convert int to string
	err := Unmarshal(suite.data, &suite.car)
	suite.Assert().Nil(err)
	suite.CarAssertions()
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseByte_DifferentType() {
	data := []byte(`{"Brand":"Toyota","Year":"2020","Used":true,"Price":"100000.50","Document":123}`)
	err := Unmarshal(data, &suite.car)
	suite.Assert().Nil(err)
	suite.CarAssertions()
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseJsonCar() {
	car := Car{
		Brand:    BrandValue,
		Year:     YearValue,
		Used:     UsedValue,
		Price:    PriceValue,
		Document: DocumentValue,
	}

	jsonData, err := json.Marshal(car)
	if err != nil {
		suite.T().Fatal(err)
	}

	err = Unmarshal(jsonData, &suite.car)

	suite.Assert().Nil(err)
	suite.CarAssertions()
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseJsonCars() {
	cars := []Car{
		{
			Brand:    BrandValue,
			Year:     YearValue,
			Used:     UsedValue,
			Price:    PriceValue,
			Document: DocumentValue,
		},
	}

	jsonData, err := json.Marshal(cars)
	if err != nil {
		suite.T().Fatal(err)
	}

	err = Unmarshal(jsonData, &suite.cars)

	suite.Assert().Nil(err)
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseMap_MissingProperty() {
	delete(suite.data, "Brand")
	err := Unmarshal(suite.data, &suite.car)
	errs := multierr.Errors(err)
	suite.Assert().Nil(err)
	suite.Assert().Len(errs, 0)
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseMap_AdditionalProperty() {
	suite.data["Additional"] = "Additional"
	suite.data["Additional2"] = 123
	err := Unmarshal(suite.data, &suite.car)
	suite.Assert().Nil(err)
	suite.CarAssertions()
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseMap_InvalidType() {
	suite.data["Year"] = "A"  // invalid type for int conversion
	suite.data["Price"] = "B" // invalid type for float64 conversion
	err := Unmarshal(suite.data, &suite.car)
	errs := multierr.Errors(err)
	suite.Assert().NotNil(err)
	suite.Assert().Len(errs, 2)
	for _, err := range errs {
		if fieldErr, ok := err.(*FieldError); ok {
			suite.Assert().True(fieldErr.IsFieldAffected())                      // use this to check if the field of the struct is affected
			suite.Assert().Contains([]string{"Year", "Price"}, fieldErr.Field()) // use this to check which field is affected
		}
	}
}

func (suite *unmarshalTestSuite) TestUnmarshal_UseUnsupportedType() {
	data := 123 // unsupported type
	err := Unmarshal(data, &suite.car)
	suite.Assert().NotNil(err)
}

func TestStruct(t *testing.T) {
	suite.Run(t, new(unmarshalTestSuite))
}
