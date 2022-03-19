package csv

import (
	"fmt"
	"go4ml.xyz/data"
	"go4ml.xyz/iokit"
	"testing"
)

var irisSource = iokit.Url("https://datahub.io/machine-learning/iris/r/iris.csv", iokit.Cache("dataset/iris.csv"))

func Test_IrisCsv_1(t *testing.T) {
	cls := data.Enumset{}
	irisCsv := Source(irisSource,
		Float32("sepallength").As("Feature1"),
		Float32("sepalwidth").As("Feature2"),
		Float32("petallength").As("Feature3"),
		Float32("petalwidth").As("Feature4"),
		Meta(cls.Integer(), "class").As("label"))

	var q data.Table
	irisCsv.MustDrain(q.Sink(), 4)

	for i := 0; i < q.Len(); i++ {
		fmt.Println(q.Row(i))
	}
}

func Test_IrisCsv_2(t *testing.T) {
	cls := data.Enumset{}
	irisCsv := Source(irisSource,
		Float32("sepal*").Group("F1"),
		Float32("petal*").Group("F2"),
		Meta(cls.Integer(), "class").As("label"))

	var q data.Table
	irisCsv.MustDrain(q.Sink(), 4)

	for i := 0; i < q.Len(); i++ {
		fmt.Println(q.Row(i))
	}
}

