package main

import (
	"fmt"
	"github.com/Edu963/ocultar/internal/pii"
)

func main() {
	input := "My name is John Doe, my email is john.doe@example.com and my phone is +1-555-0199."
	eng := pii.NewRefinery()
	
	fmt.Println("Input:", input)
	
	detections := eng.Scan(input)
	fmt.Printf("Found %d detections\n", len(detections))
	for _, d := range detections {
		fmt.Printf("- [%s] %s (%s)\n", d.Entity, d.Value, d.Location)
	}
	
	redacted, _ := eng.Redact(input, func(d pii.DetectionResult) (string, error) {
		return "[REDACTED]", nil
	})
	fmt.Println("Redacted:", redacted)
}
