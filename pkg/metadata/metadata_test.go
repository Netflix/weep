package metadata

import (
	"os"
	"os/user"
	"testing"
	"time"
)

func TestGetInstanceInfo(t *testing.T) {
	expectedHostname, _ := os.Hostname()
	u, _ := user.Current()
	expectedUsername := u.Username
	timeDiff := time.Second * 5
	certCreationTime = time.Now().Add(-timeDiff)
	certFingerprint = "👉🖨"
	result := GetInstanceInfo()
	if result.Hostname != expectedHostname {
		t.Errorf("hostname: expected %s, got %s", expectedHostname, result.Hostname)
	}
	if result.Username != expectedUsername {
		t.Errorf("username: expected %s, got %s", expectedUsername, result.Username)
	}
	if !(result.CertAgeSeconds >= 4 && result.CertAgeSeconds <= 6) {
		t.Errorf("cert age seconds: expected 4 <= x <= 6, got %d", result.CertAgeSeconds)
	}
	if result.CertFingerprintSHA256 != certFingerprint {
		t.Errorf("cert fingerprint: expected %s, got %s", certFingerprint, result.CertFingerprintSHA256)
	}
	if result.WeepVersion != Version {
		t.Errorf("weep version: expected %s, got %s", Version, result.WeepVersion)
	}
	if result.WeepMethod != weepMethod {
		t.Errorf("weep method: expected %s, got %s", weepMethod, result.WeepMethod)
	}
}

func TestElapsedSeconds(t *testing.T) {
	cases := []struct {
		Description string
		StartTime   time.Time
		EndTime     time.Time
		Expected    int
	}{
		{
			Description: "happy path, start before end",
			StartTime:   time.Unix(50000, 0),
			EndTime:     time.Unix(50005, 0),
			Expected:    5,
		},
		{
			Description: "start before end with ns",
			StartTime:   time.Unix(50000, 123),
			EndTime:     time.Unix(50005, 321),
			Expected:    5,
		},
		{
			Description: "end before start",
			StartTime:   time.Unix(50005, 0),
			EndTime:     time.Unix(50000, 0),
			Expected:    -5,
		},
		{
			Description: "zero start time",
			StartTime:   time.Time{},
			EndTime:     time.Unix(50000, 0),
			Expected:    0,
		},
		{
			Description: "zero end time",
			StartTime:   time.Time{},
			EndTime:     time.Unix(0, 0),
			Expected:    0,
		},
	}
	for i, tc := range cases {
		t.Logf("test case %d: %s", i, tc.Description)
		result := elapsedSeconds(tc.StartTime, tc.EndTime)
		if result != tc.Expected {
			t.Errorf("%s failed: expected %d, got %d", tc.Description, tc.Expected, result)
		}
	}
}

func TestSetCertInfo(t *testing.T) {
	creationTime := time.Now()
	fingerprint := "👉🖨"
	SetCertInfo(creationTime, fingerprint)
	if certCreationTime != creationTime {
		t.Errorf("set cert creation time failed: expected %v, got %v", creationTime, certCreationTime)
	}
	if certFingerprint != fingerprint {
		t.Errorf("set cert fingerprint failed: expected %s, got %s", fingerprint, certFingerprint)
	}
}

func TestHostname(t *testing.T) {
	expected, _ := os.Hostname()
	result := hostname()
	if result != expected {
		t.Errorf("hostname failed: expected %s, got %s", expected, result)
	}
}

func TestUsername(t *testing.T) {
	u, _ := user.Current()
	expected := u.Username
	result := username()
	if result != expected {
		t.Errorf("username failed: expected %s, got %s", expected, result)
	}
}

func TestSetWeepMethod(t *testing.T) {
	expected := "cool-command"
	SetWeepMethod(expected)
	if weepMethod != expected {
		t.Errorf("weep method failed: expected %s, got %s", expected, weepMethod)
	}
}

func TestStartupTime(t *testing.T) {
	weepStartupTime = time.Date(2020, 1, 2, 3, 45, 67, 0, time.UTC)
	expected := "2020-01-02T03:46:07Z"
	result := StartupTime()
	if result != expected {
		t.Errorf("startup time failed: expected %s, got %s", expected, result)
	}
}
