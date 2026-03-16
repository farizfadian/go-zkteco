// Example usage of go-zkteco library
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/farizfadian/go-zkteco"
)

func main() {
	// Get device IP from command line or environment
	deviceIP := os.Getenv("ZKTECO_IP")
	if deviceIP == "" {
		if len(os.Args) > 1 {
			deviceIP = os.Args[1]
		} else {
			deviceIP = "192.168.1.201"
		}
	}

	fmt.Printf("Connecting to ZKTeco device at %s...\n", deviceIP)

	// Connect to device
	device, err := zkteco.Connect(deviceIP,
		zkteco.WithTimeout(10*time.Second),
		zkteco.WithRetry(3, time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer device.Disconnect()

	fmt.Println("Connected!")
	fmt.Println()

	// Get device info
	fmt.Println("=== Device Info ===")
	info, err := device.GetDeviceInfo()
	if err != nil {
		log.Printf("Warning: Could not get device info: %v", err)
	} else {
		fmt.Printf("Serial Number : %s\n", info.SerialNumber)
		fmt.Printf("Device Name   : %s\n", info.DeviceName)
		fmt.Printf("Platform      : %s\n", info.Platform)
		fmt.Printf("Firmware      : %s\n", info.FirmwareVersion)
	}
	fmt.Println()

	// Get device time
	fmt.Println("=== Device Time ===")
	deviceTime, err := device.GetTime()
	if err != nil {
		log.Printf("Warning: Could not get device time: %v", err)
	} else {
		fmt.Printf("Device Time   : %s\n", deviceTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("Local Time    : %s\n", time.Now().Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	// Get users
	fmt.Println("=== Users ===")
	users, err := device.GetUsers()
	if err != nil {
		log.Printf("Warning: Could not get users: %v", err)
	} else {
		fmt.Printf("Total users: %d\n", len(users))
		for i, user := range users {
			if i >= 10 {
				fmt.Printf("  ... and %d more\n", len(users)-10)
				break
			}
			fmt.Printf("  ID: %d, Name: %s, Privilege: %s\n",
				user.UserID, user.Name, user.PrivilegeString())
		}
	}
	fmt.Println()

	// Get attendance logs
	fmt.Println("=== Attendance Logs ===")
	logs, err := device.GetAttendance()
	if err != nil {
		log.Printf("Warning: Could not get attendance logs: %v", err)
	} else {
		fmt.Printf("Total records: %d\n", len(logs))

		// Show last 10 records
		start := 0
		if len(logs) > 10 {
			start = len(logs) - 10
			fmt.Printf("Showing last 10 records:\n")
		}

		for _, record := range logs[start:] {
			fmt.Printf("  User %d: %s (%s via %s)\n",
				record.UserID,
				record.Time.Format("2006-01-02 15:04:05"),
				record.StateString(),
				record.VerifyTypeString())
		}
	}
	fmt.Println()

	fmt.Println("Done!")
}
