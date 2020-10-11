package web_test

import (
	"encoding/json"
	"fmt"
	mock_config "github.com/analogj/scrutiny/webapp/backend/pkg/config/mock"
	dbModels "github.com/analogj/scrutiny/webapp/backend/pkg/models/db"
	"github.com/analogj/scrutiny/webapp/backend/pkg/web"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

func TestHealthRoute(t *testing.T) {
	//setup
	parentPath, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("web.database.location").Return(path.Join(parentPath, "scrutiny_test.db")).AnyTimes()
	fakeConfig.EXPECT().GetString("web.src.frontend.path").Return(parentPath).AnyTimes()

	ae := web.AppEngine{
		Config: fakeConfig,
	}

	router := ae.Setup(logrus.New())

	//test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	router.ServeHTTP(w, req)

	//assert
	require.Equal(t, 200, w.Code)
	require.Equal(t, "{\"success\":true}", w.Body.String())
}

func TestRegisterDevicesRoute(t *testing.T) {
	//setup
	parentPath, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("web.database.location").Return(path.Join(parentPath, "scrutiny_test.db")).AnyTimes()
	fakeConfig.EXPECT().GetString("web.src.frontend.path").Return(parentPath).AnyTimes()
	ae := web.AppEngine{
		Config: fakeConfig,
	}
	router := ae.Setup(logrus.New())
	file, err := os.Open("testdata/register-devices-req.json")
	require.NoError(t, err)

	//test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/devices/register", file)
	router.ServeHTTP(w, req)

	//assert
	require.Equal(t, 200, w.Code)
}

func TestUploadDeviceMetricsRoute(t *testing.T) {
	//setup
	parentPath, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("web.database.location").AnyTimes().Return(path.Join(parentPath, "scrutiny_test.db"))
	fakeConfig.EXPECT().GetString("web.src.frontend.path").AnyTimes().Return(parentPath)
	ae := web.AppEngine{
		Config: fakeConfig,
	}
	router := ae.Setup(logrus.New())
	devicesfile, err := os.Open("testdata/register-devices-single-req.json")
	require.NoError(t, err)

	metricsfile, err := os.Open("testdata/upload-device-metrics-req.json")
	require.NoError(t, err)

	//test
	wr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/devices/register", devicesfile)
	router.ServeHTTP(wr, req)
	require.Equal(t, 200, wr.Code)

	mr := httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/device/0x5000cca264eb01d7/smart", metricsfile)
	router.ServeHTTP(mr, req)
	require.Equal(t, 200, mr.Code)

	//assert
}

func TestPopulateMultiple(t *testing.T) {
	//setup
	parentPath, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	//fakeConfig.EXPECT().GetString("web.database.location").AnyTimes().Return("testdata/scrutiny_test.db")
	fakeConfig.EXPECT().GetStringSlice("notify.urls").Return([]string{}).AnyTimes()
	fakeConfig.EXPECT().GetString("web.database.location").AnyTimes().Return(path.Join(parentPath, "scrutiny_test.db"))
	fakeConfig.EXPECT().GetString("web.src.frontend.path").AnyTimes().Return(parentPath)
	ae := web.AppEngine{
		Config: fakeConfig,
	}
	router := ae.Setup(logrus.New())
	devicesfile, err := os.Open("testdata/register-devices-req.json")
	require.NoError(t, err)

	metricsfile, err := os.Open("../models/testdata/smart-ata.json")
	require.NoError(t, err)
	failfile, err := os.Open("../models/testdata/smart-fail2.json")
	require.NoError(t, err)
	nvmefile, err := os.Open("../models/testdata/smart-nvme.json")
	require.NoError(t, err)
	scsifile, err := os.Open("../models/testdata/smart-scsi.json")
	require.NoError(t, err)
	scsi2file, err := os.Open("../models/testdata/smart-scsi2.json")
	require.NoError(t, err)

	//test
	wr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/devices/register", devicesfile)
	router.ServeHTTP(wr, req)
	require.Equal(t, 200, wr.Code)

	mr := httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/device/0x5000cca264eb01d7/smart", metricsfile)
	router.ServeHTTP(mr, req)
	require.Equal(t, 200, mr.Code)

	fr := httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/device/0x5000cca264ec3183/smart", failfile)
	router.ServeHTTP(fr, req)
	require.Equal(t, 200, fr.Code)

	nr := httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/device/0x5002538e40a22954/smart", nvmefile)
	router.ServeHTTP(nr, req)
	require.Equal(t, 200, nr.Code)

	sr := httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/device/0x5000cca252c859cc/smart", scsifile)
	router.ServeHTTP(sr, req)
	require.Equal(t, 200, sr.Code)

	s2r := httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/device/0x5000cca264ebc248/smart", scsi2file)
	router.ServeHTTP(s2r, req)
	require.Equal(t, 200, s2r.Code)

	//assert
}

func TestPopulateMultiple_Bulk(t *testing.T) {
	//setup
	parentPath, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	//fakeConfig.EXPECT().GetString("web.database.location").AnyTimes().Return("testdata/scrutiny_test.db")
	fakeConfig.EXPECT().GetStringSlice("notify.urls").Return([]string{}).AnyTimes()
	fakeConfig.EXPECT().GetString("web.database.location").AnyTimes().Return(path.Join(parentPath, "scrutiny_test.db"))
	fakeConfig.EXPECT().GetString("web.src.frontend.path").AnyTimes().Return(parentPath)
	ae := web.AppEngine{
		Config: fakeConfig,
	}
	router := ae.Setup(logrus.New())
	devicesfile, err := os.Open("testdata/bulk/register-devices-req.json")
	require.NoError(t, err)

	devices := []string{
		"0x5000c500ad8fbe67",
		"0x5000c500ad85a42f",
		"0x5000c500ad85889f",
		"0x5000c500aebeb223",
		"0x5000c5007f7a0153",
		"0x5000c500773a1fcf",
		"0x5000c500773be1cb",
		"0x5000c500773c5c37",
		"0x5000c500773c7f47",
		"0x5000c500773c9abf",
		"0x5000c500773c64e7",
		"0x5000c500773c6013",
		"0x5000c500773c7477",
		"0x5000c500773cbb2f",
		"0x5000c500773d4bd3",
		"0x5000c500773d35a7",
		"0x5000c500773da3cb",
		"0x5000c500773dabab",
		"0x5000c5007726bc6b",
		"0x5000c50077263fcf",
		"0x5000c50077264aaf",
		"0x5000c50077267db7",
		"0x5000c50077269e43",
		"0x5000c500772076f3",
		"0x5000c500772351df",
		"0x5000c500772673c3",
		"0x5000c5007720819b",
		"0x5000c5007726653f",
		"0x5000c50077267483",
		"0x5000c50084036983",
		"0x50000395a8107f1c",
		"0x50000395a8138a7c",
		"0x50025388a06e63a2",
		"0x50025388a0682aae",
		"0x50025388a068237f",
		"0x500253887001d5e2",
		"0x500253887001d61f",
		"0x5002538840073ad9",
	}

	//
	//
	//metricsfile, err := os.Open("../models/testdata/smart-ata.json")
	//require.NoError(t, err)
	//failfile, err := os.Open("../models/testdata/smart-fail2.json")
	//require.NoError(t, err)
	//nvmefile, err := os.Open("../models/testdata/smart-nvme.json")
	//require.NoError(t, err)
	//scsifile, err := os.Open("../models/testdata/smart-scsi.json")
	//require.NoError(t, err)
	//scsi2file, err := os.Open("../models/testdata/smart-scsi2.json")
	//require.NoError(t, err)

	//test
	wr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/devices/register", devicesfile)
	router.ServeHTTP(wr, req)
	require.Equal(t, 200, wr.Code)

	for _, device := range devices {
		deviceFile, err := os.Open(fmt.Sprintf("testdata/bulk/%s.json", device))
		require.NoError(t, err)

		mr := httptest.NewRecorder()
		req, _ = http.NewRequest("POST", fmt.Sprintf("/api/device/%s/smart", device), deviceFile)
		router.ServeHTTP(mr, req)
		require.Equal(t, 200, mr.Code)
	}
	//assert
}

//TODO: this test should use a recorded request/response playback.
//func TestSendTestNotificationRoute(t *testing.T) {
//	//setup
//	parentPath, _ := ioutil.TempDir("", "")
//	defer os.RemoveAll(parentPath)
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//	fakeConfig := mock_config.NewMockInterface(mockCtrl)
//	fakeConfig.EXPECT().GetString("web.database.location").AnyTimes().Return(path.Join(parentPath, "scrutiny_test.db"))
//	fakeConfig.EXPECT().GetString("web.src.frontend.path").AnyTimes().Return(parentPath)
//	fakeConfig.EXPECT().GetStringSlice("notify.urls").AnyTimes().Return([]string{"https://scrutiny.requestcatcher.com/test"})
//	ae := web.AppEngine{
//		Config: fakeConfig,
//	}
//	router := ae.Setup(logrus.New())
//
//	//test
//	wr := httptest.NewRecorder()
//	req, _ := http.NewRequest("POST", "/api/health/notify", strings.NewReader("{}"))
//	router.ServeHTTP(wr, req)
//
//	//assert
//	require.Equal(t, 200, wr.Code)
//}

func TestSendTestNotificationRoute_WebhookFailure(t *testing.T) {
	//setup
	parentPath, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("web.database.location").AnyTimes().Return(path.Join(parentPath, "scrutiny_test.db"))
	fakeConfig.EXPECT().GetString("web.src.frontend.path").AnyTimes().Return(parentPath)
	fakeConfig.EXPECT().GetStringSlice("notify.urls").AnyTimes().Return([]string{"https://unroutable.domain.example.asdfghj"})
	ae := web.AppEngine{
		Config: fakeConfig,
	}
	router := ae.Setup(logrus.New())

	//test
	wr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/health/notify", strings.NewReader("{}"))
	router.ServeHTTP(wr, req)

	//assert
	require.Equal(t, 500, wr.Code)
}

func TestSendTestNotificationRoute_ScriptFailure(t *testing.T) {
	//setup
	parentPath, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("web.database.location").AnyTimes().Return(path.Join(parentPath, "scrutiny_test.db"))
	fakeConfig.EXPECT().GetString("web.src.frontend.path").AnyTimes().Return(parentPath)
	fakeConfig.EXPECT().GetStringSlice("notify.urls").AnyTimes().Return([]string{"script:///missing/path/on/disk"})
	ae := web.AppEngine{
		Config: fakeConfig,
	}
	router := ae.Setup(logrus.New())

	//test
	wr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/health/notify", strings.NewReader("{}"))
	router.ServeHTTP(wr, req)

	//assert
	require.Equal(t, 500, wr.Code)
}

func TestSendTestNotificationRoute_ShoutrrrFailure(t *testing.T) {
	//setup
	parentPath, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("web.database.location").AnyTimes().Return(path.Join(parentPath, "scrutiny_test.db"))
	fakeConfig.EXPECT().GetString("web.src.frontend.path").AnyTimes().Return(parentPath)
	fakeConfig.EXPECT().GetStringSlice("notify.urls").AnyTimes().Return([]string{"discord://invalidtoken@channel"})
	ae := web.AppEngine{
		Config: fakeConfig,
	}
	router := ae.Setup(logrus.New())

	//test
	wr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/health/notify", strings.NewReader("{}"))
	router.ServeHTTP(wr, req)

	//assert
	require.Equal(t, 500, wr.Code)
}

func TestGetDevicesSummaryRoute_Nvme(t *testing.T) {
	//setup
	parentPath, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("web.database.location").AnyTimes().Return(path.Join(parentPath, "scrutiny_test.db"))
	fakeConfig.EXPECT().GetString("web.src.frontend.path").AnyTimes().Return(parentPath)
	ae := web.AppEngine{
		Config: fakeConfig,
	}
	router := ae.Setup(logrus.New())
	devicesfile, err := os.Open("testdata/register-devices-req-2.json")
	require.NoError(t, err)

	metricsfile, err := os.Open("../models/testdata/smart-nvme2.json")
	require.NoError(t, err)

	//test
	wr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/devices/register", devicesfile)
	router.ServeHTTP(wr, req)
	require.Equal(t, 200, wr.Code)

	mr := httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/device/a4c8e8ed-11a0-4c97-9bba-306440f1b944/smart", metricsfile)
	router.ServeHTTP(mr, req)
	require.Equal(t, 200, mr.Code)

	sr := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/summary", nil)
	router.ServeHTTP(sr, req)
	require.Equal(t, 200, sr.Code)
	var device dbModels.DeviceWrapper
	json.Unmarshal(sr.Body.Bytes(), &device)

	//assert
	require.Equal(t, "a4c8e8ed-11a0-4c97-9bba-306440f1b944", device.Data[0].WWN)
	require.Equal(t, "passed", device.Data[0].SmartResults[0].SmartStatus)
}
