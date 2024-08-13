package main

import (
	."github.com/onsi/gomega"
	"testing"
	."github.com/onsi/ginkgo/v2"

)
const chicagoLatitude= 41.8781
const chicagoLongitude= -87.6298
// This function uses ginkgo and gomega to test the getCoordinates function to check if the city New York
// does not have the same coordinates as another city, in this case chicago
func TestCoordinates(t *testing.T){
	gomega.RegisterFailHandler(ginkgo.Fail)
	city := []string{"New York"}
	longitude,latitude,err := getCoordinates(city[0])
	Expect(longitude).To(NotEqual(chicagoLongitude))
	Expect(latitude).To(NotEqual(chicagoLatitude))
	Expect(err).ToNot(BeNil())
}

