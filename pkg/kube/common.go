package kube

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora/v3"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	warning_threshold  = 90.00
	critical_threshold = 95.00
)

//calcPercentage
func calcPercentage(dividend, divisor int64) float64 {
	if divisor > 0 {
		value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(float64(dividend)/float64(divisor)*100)), 64)
		return value
	}
	return float64(0)
}

type MemoryResource struct {
	*resource.Quantity
}

//NewMemoryResource
func NewMemoryResource(value int64) *MemoryResource {
	return &MemoryResource{resource.NewQuantity(value, resource.BinarySI)}
}

//calcPercentage
func (r *MemoryResource) calcPercentage(divisor *resource.Quantity) float64 {
	return calcPercentage(r.Value(), divisor.Value())
}

func (r *MemoryResource) String() string {
	// XXX: Support more units
	return fmt.Sprintf("%vMi", r.Value()/(1024*1024))
}

//ToQuantity
func (r *MemoryResource) ToQuantity() *resource.Quantity {
	return resource.NewQuantity(r.Value(), resource.BinarySI)
}

type CpuResource struct {
	*resource.Quantity
}

//NewCpuResource
func NewCpuResource(value int64) *CpuResource {
	r := resource.NewMilliQuantity(value, resource.DecimalSI)
	return &CpuResource{r}
}

//String
func (r *CpuResource) String() string {
	// XXX: Support more units
	return fmt.Sprintf("%vm", r.MilliValue())
}

//intToString int转string
func intToString(a int) string {
	str := strconv.Itoa(a)
	return str
}

//float64ToString float64转string
func float64ToString(s float64) string {
	//return strconv.FormatFloat(s, 'G', -1, 32)
	return fmt.Sprintf("%v%%", strconv.FormatFloat(s, 'G', -1, 32))

}

//int64ToString int64转string
func int64ToString(a int64) string {
	str := strconv.FormatInt(a, 10)
	return str
}

//StringTofloat64
func stringTofloat64(a string) float64 {
	value, _ := strconv.ParseFloat(a, 64)
	return value
}

//calcPercentage
func (r *CpuResource) calcPercentage(divisor *resource.Quantity) float64 {
	return calcPercentage(r.MilliValue(), divisor.MilliValue())
}

//ToQuantity
func (r *CpuResource) ToQuantity() *resource.Quantity {
	return resource.NewMilliQuantity(r.MilliValue(), resource.DecimalSI)
}

//FieldString
func FieldString(str string) float64 {
	switch {
	case strings.Contains(str, "%"):
		str1 := strings.Split(str, "%")
		value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", stringTofloat64(str1[0])), 64)
		return value
	case strings.Contains(str, "Mi"):
		str1 := strings.Split(str, "Mi")
		value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", stringTofloat64(str1[0])), 64)
		return value
	case strings.Contains(str, "m"):
		str1 := strings.Split(str, "m")
		value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", stringTofloat64(str1[0])), 64)
		return value
	default:
		return float64(0)
	}
}

//Compare
func ExceedsCompare(a string) string {
	if FieldString(a) > float64(critical_threshold) {
		return redColor(a)
	} else if FieldString(a) > float64(warning_threshold) {
		return yellowColor(a)
	} else {
		return a
	}
}

func redColor(s string) string {
	return fmt.Sprintf("%s", aurora.Red(s))
}

func yellowColor(s string) string {
	return fmt.Sprintf("%s", aurora.Yellow(s))
}
