#include <Arduino.h>

#include <TFT_eSPI.h> 
#include <SPI.h>
#include "WiFi.h"
#include <Wire.h>
#include <Button2.h>
#include "esp_adc_cal.h"
#include <TinyGPSPlus.h>
#include <Sd.h>
#include "mpu9250.h"
#include "SensorFusion.h"
#include <Adafruit_Sensor.h>
#include <DHT.h>
#include <DHT_U.h>

#define DHTPIN 25
#define DHTTYPE    DHT11

DHT_Unified dht(DHTPIN, DHTTYPE);
uint32_t delayMS;

#ifndef TFT_DISPOFF
#define TFT_DISPOFF 0x28
#endif

#ifndef TFT_SLPIN
#define TFT_SLPIN   0x10
#endif

#define ADC_EN          14
#define ADC_PIN         34
#define BUTTON_1        35
#define BUTTON_2        0

// MPU9250 IMU(Wire,0x68);

// SF fusion;

#include "FS.h"
#include <LittleFS.h>
#define FORMAT_LITTLEFS_IF_FAILED false

#include <soc/efuse_reg.h>
#include "esp_efuse.h"

#include <SparkFunMPU9250-DMP.h>
#include <HTTPClient.h>

MPU9250_DMP imu;

float gx, gy, gz, ax, ay, az, mx, my, mz, temp;
float pitch, roll, yaw;
float deltat;

float roll_offset = 0.0;
float pitch_offset = 0.0;
float yaw_offset = 0.0;

TFT_eSPI tft = TFT_eSPI(135, 240); // Invoke custom library
Button2 btn1(BUTTON_1);
Button2 btn2(BUTTON_2);

bool flag = false;
bool clean = false;

char buff[512];
int vref = 1100;
int btnCick = false;

TinyGPSPlus gps;
unsigned long last = 0UL;

String SSID = "trackernet";
String PASSWORD = "TrackerNet@2023"; 

String deviceID = "device-001";
String password = "password01";

HTTPClient http;

float int_temp = 0.0;

void wifi_scan()
{
    tft.setTextColor(TFT_GREEN, TFT_BLACK);
    tft.fillScreen(TFT_BLACK);
    tft.setTextDatum(MC_DATUM);
    tft.setTextSize(1);

    tft.drawString("Scan Network", tft.width() / 2, tft.height() / 2);

    WiFi.mode(WIFI_STA);
    WiFi.disconnect();
    delay(100);

    int16_t n = WiFi.scanNetworks();
    tft.fillScreen(TFT_BLACK);
    if (n == 0) {
        tft.drawString("no networks found", tft.width() / 2, tft.height() / 2);
    } else {
        tft.setTextDatum(TL_DATUM);
        tft.setCursor(0, 0);
        Serial.printf("Found %d net\n", n);
        for (int i = 0; i < n; ++i) {
            sprintf(buff,
                    "[%d]:%s(%d)",
                    i + 1,
                    WiFi.SSID(i).c_str(),
                    WiFi.RSSI(i));
            tft.println(buff);
        }
    }
    WiFi.mode(WIFI_OFF);
	dht.begin();
  sensor_t sensor;
  dht.temperature().getSensor(&sensor);
  delayMS = sensor.min_delay / 1000;
}

//! Long time delay, it is recommended to use shallow sleep, which can effectively reduce the current consumption
void espDelay(int ms)
{   
    esp_sleep_enable_timer_wakeup(ms * 1000);
    esp_sleep_pd_config(ESP_PD_DOMAIN_RTC_PERIPH,ESP_PD_OPTION_ON);
    esp_light_sleep_start();
}

struct data_sample
{
	float max_ax = 0.0;
	float max_ay = 0.0;
	float max_az = 0.0;
	float internal_temp = 0.0;
	float external_temp = 0.0;
	float max_pitch = 0.0;
	float max_roll = 0.0;
	float max_yaw = 0.0;
	float lat = 0.0;
	float lng = 0.0;
	float alt = 0.0;
	float hum = 0.0;
	int sats = 0;
	String timestamp;
};

struct data_sample data_buffer[31];
int buff_index = 0;

void setup()
{
    Serial.begin(115200);
    Serial.println("Start");
    Serial2.begin(9600, SERIAL_8N1, 26, 25);
    
    tft.init();
    tft.setRotation(0);
    tft.fillScreen(TFT_BLACK);
    tft.setTextSize(2);
    tft.setTextColor(TFT_WHITE);
    tft.setCursor(0, 0);
    tft.setTextDatum(MC_DATUM);
    tft.setTextSize(1);

    if (TFT_BL > 0) { // TFT_BL has been set in the TFT_eSPI library in the User Setup file TTGO_T_Display.h
         pinMode(TFT_BL, OUTPUT); // Set backlight pin to output mode
         digitalWrite(TFT_BL, TFT_BACKLIGHT_ON); // Turn backlight on. TFT_BACKLIGHT_ON has been set in the TFT_eSPI library in the User Setup file TTGO_T_Display.h
    }

	tft.fillRect(0, 0, 135, 120, TFT_WHITE);
	tft.fillRect(0, 120, 135, 120, TFT_WHITE);

	// Show GPS Fix status on first box
	tft.setTextColor(TFT_WHITE, TFT_BLACK);
	tft.setTextDatum(TL_DATUM);
	tft.drawString("Fix Status:", 2, 2);
	tft.setTextColor(TFT_YELLOW, TFT_BLACK);
	tft.drawString("Waiting for GPS", 2, 12);

	// Show orientation and acceleration data on second box
	tft.setTextColor(TFT_WHITE, TFT_BLACK);
	tft.setTextDatum(TL_DATUM);
	tft.drawString("Orientation:", 2, 122);
	tft.setTextColor(TFT_YELLOW, TFT_BLACK);
	tft.drawString("Waiting for MPU9250", 2, 132);

    Wire.begin();

	if(imu.begin() != INV_SUCCESS){
		Serial.println("MPU9250 does not respond");
	}
	else{
		Serial.println("MPU9250 is connected");
	}

	imu.dmpBegin(DMP_FEATURE_6X_LP_QUAT | // Enable 6-axis quat
				DMP_FEATURE_GYRO_CAL, // Use gyro calibration
				10); // Set DMP FIFO rate to 10 Hz



    esp_adc_cal_characteristics_t adc_chars;
    esp_adc_cal_value_t val_type = esp_adc_cal_characterize((adc_unit_t)ADC_UNIT_1, (adc_atten_t)ADC1_CHANNEL_6, (adc_bits_width_t)ADC_WIDTH_BIT_12, 1100, &adc_chars);
    //Check type of calibration value used to characterize ADC
    if (val_type == ESP_ADC_CAL_VAL_EFUSE_VREF) {
        Serial.printf("eFuse Vref:%u mV", adc_chars.vref);
        vref = adc_chars.vref;
    } else if (val_type == ESP_ADC_CAL_VAL_EFUSE_TP) {
        Serial.printf("Two Point --> coeff_a:%umV coeff_b:%umV\n", adc_chars.coeff_a, adc_chars.coeff_b);
    } else {
        Serial.println("Default Vref: 1100mV");
    }
	// Initialize LittleFS if not already done
	if(!LittleFS.begin(FORMAT_LITTLEFS_IF_FAILED)){
	Serial.println("LittleFS Mount Failed");
	return;
	}
}



void loop()
{
    while (Serial2.available() > 0) {
		gps.encode(Serial2.read());
					    Serial.print(F("DIAGS      Chars="));
    Serial.print(gps.charsProcessed());
    Serial.print(F(" Sentences-with-Fix="));
    Serial.print(gps.sentencesWithFix());
    Serial.print(F(" Failed-checksum="));
    Serial.print(gps.failedChecksum());
    Serial.print(F(" Passed-checksum="));
    Serial.println(gps.passedChecksum());

	}

	// If disconnected from WiFi, try to reconnect
	if (WiFi.status() != WL_CONNECTED) {
		WiFi.begin(SSID, PASSWORD);
		while (WiFi.status() != WL_CONNECTED) {
			delay(1000);
			Serial.println("Connecting to WiFi..");
		}
		Serial.println("Connected to the WiFi network");
	}

	// Wait for GPS to get a fix
	// while (!gps.location.isValid()) {
	// 	if (millis() - last > 1000) {
	// 		last = millis();
	// 		Serial.println("Waiting for GPS fix..");
	// 		Serial.println("Number of satellites: " + String(gps.satellites.value()));

	// 	}
	//     while (Serial2.available() > 0) {
	// 		gps.encode(Serial2.read());
	// 	}
	// }

	// Read internal temperature
    sensors_event_t event;
	dht.temperature().getEvent(&event);
	if (isnan(event.temperature)) {
		Serial.println("Error reading temperature!");
	} else {
		Serial.print("Temperature: ");
		Serial.print(event.temperature);
		Serial.println(" *C");
		int_temp = event.temperature;
	}

	File file = LittleFS.open("/initialized.txt", "rw+");
	if (!file) {
		Serial.println("Failed to open file for reading");
		return;
	}
	file.seek(0);
	String line = file.readStringUntil('\n');
	Serial.println(line);
	if (line == "true") {
		Serial.println("Device is already initialized");
	} else {
		Serial.println("Device is not initialized");
		while (digitalRead(BUTTON_1) == HIGH) {
			delay(1000);
			Serial.println("Waiting for button press..");
			file = LittleFS.open("/initialized.txt", "r");
			if (!file) {
				Serial.println("Failed to open file for reading");
				return;
			}
			file.seek(0);
			line = file.readStringUntil('\n');
			Serial.println(line);
			file.close();
			tft.setTextColor(TFT_WHITE, TFT_BLACK);
			tft.setTextDatum(TL_DATUM);
			tft.drawString("Press Button 1", 2, 2);
			tft.setTextColor(TFT_YELLOW, TFT_BLACK);
			tft.drawString("Waiting for button press..", 2, 12);
		}
		http.begin("http://94.100.26.221:8001/api/init");
		http.setTimeout(100000);
		http.addHeader("Content-Type", "application/x-www-form-urlencoded");
		//String payload = "{\"deviceID\": \"" + deviceID + "\", \"password\": \"" + password + "\", \"capabilities\": [\"accelerometer\", \"gyroscope\", \"magnetometer\", \"gps\", \"temperature\", \"orientation\"]}";
		String payload = "deviceID=" + deviceID + "&password=" + password + "&capabilities=accelerometer,gyroscope,magnetometer,gps,temperature,orientation";
		Serial.println(payload);
		int httpResponseCode = http.POST(payload);
		if (httpResponseCode > 0) {
			Serial.print("HTTP Response code: ");
			Serial.println(httpResponseCode);
			String response = http.getString();
			Serial.println(response);
		} else {
			Serial.print("Error code: ");
			Serial.println(httpResponseCode);
		}
		http.end();
			// Write to LittleFS that device is initialized
		file = LittleFS.open("/initialized.txt", "w+");
		if (!file) {
			Serial.println("Failed to open file for writing");
			return;
		}
		file.seek(0);
		if (file.print("true")) {
			Serial.println("File written");
		} else {
			Serial.println("Write failed");
		}
		file.close();
	}
	file.close();

	// read raw accel/gyro/mag measurements from device
	// IMU.readSensor();

	// ax = IMU.getAccelX_mss();
	// ay = IMU.getAccelY_mss();
	// az = IMU.getAccelZ_mss();
	// gx = IMU.getGyroX_rads();
	// gy = IMU.getGyroY_rads();
	// gz = IMU.getGyroZ_rads();
	// mx = IMU.getMagX_uT();
	// my = IMU.getMagY_uT();
	// mz = IMU.getMagZ_uT();
	// temp = IMU.getTemperature_C();

	// deltat = fusion.deltatUpdate();
	// //fusion.MahonyUpdate(gx, gy, gz, ax, ay, az, mx, my, mz, deltat);  //mahony is suggested if there isn't the mag
	// fusion.MadgwickUpdate(gx, gy, gz, ax, ay, az, mx, my, mz, deltat);  //else use the magwick

	// roll = fusion.getRoll();
	// pitch = fusion.getPitch();
	// yaw = fusion.getYaw();

	if ( imu.fifoAvailable() )
	{
		// Use dmpUpdateFifo to update the ax, gx, mx, etc. values
		if ( imu.dmpUpdateFifo() == INV_SUCCESS)
		{
		// computeEulerAngles can be used -- after updating the
		// quaternion values -- to estimate roll, pitch, and yaw
		imu.computeEulerAngles();
		float q0 = imu.calcQuat(imu.qw);
		float q1 = imu.calcQuat(imu.qx);
		float q2 = imu.calcQuat(imu.qy);
		float q3 = imu.calcQuat(imu.qz);
		imu.computeEulerAngles(true);
		roll = imu.roll;
		pitch = imu.pitch;
		yaw = imu.yaw;
		imu.update(UPDATE_TEMP | UPDATE_ACCEL | UPDATE_GYRO | UPDATE_COMPASS);
		ax = imu.calcAccel(imu.ax);
		ay = imu.calcAccel(imu.ay);
		az = imu.calcAccel(imu.az);
		imu.update(UPDATE_TEMP);
		temp = imu.updateTemperature();
		if (roll > 180) {
			roll = roll - 360;
		}
		if (pitch > 180) {
			pitch = pitch - 360;
		}
		if (yaw > 180) {
			yaw = yaw - 360;
		}
		}
	}


	if (digitalRead(BUTTON_2) == LOW) {
		roll_offset = roll;
		pitch_offset = pitch;
		yaw_offset = yaw;
	}

	roll = roll - roll_offset;
	pitch = pitch - pitch_offset;
	yaw = yaw - yaw_offset;

	if (roll > 180) {
		roll = roll - 360;
	}
	if (pitch > 180) {
		pitch = pitch - 360;
	}
	if (yaw > 180) {
		yaw = yaw - 360;
	}
	if (roll < -180) {
		roll = roll + 360;
	}
	if (pitch < -180) {
		pitch = pitch + 360;
	}
	if (yaw < -180) {
		yaw = yaw + 360;
	}

	data_buffer[buff_index].max_ax = max(abs(data_buffer[buff_index].max_ax), (ax));
	data_buffer[buff_index].max_ay = max(abs(data_buffer[buff_index].max_ay), abs(ay));
	data_buffer[buff_index].max_az = max(abs(data_buffer[buff_index].max_az), abs(az));
	if(data_buffer[buff_index].external_temp == 0.0){
		data_buffer[buff_index].external_temp = temp;
	}
	else{
		data_buffer[buff_index].external_temp = (data_buffer[buff_index].external_temp + temp)/2;
	}
	if(data_buffer[buff_index].internal_temp == 0.0){
		data_buffer[buff_index].internal_temp = int_temp;
	}
	else{
		data_buffer[buff_index].internal_temp = (data_buffer[buff_index].internal_temp + int_temp)/2;
	}
	data_buffer[buff_index].max_pitch = max(abs(data_buffer[buff_index].max_pitch), abs(pitch));
	data_buffer[buff_index].max_roll = max(abs(data_buffer[buff_index].max_roll), abs(roll));
	data_buffer[buff_index].max_yaw = max(abs(data_buffer[buff_index].max_yaw), abs(yaw));


	// Update the display
	if (millis() - last > 1000) {
		last = millis();
		tft.setTextColor(TFT_WHITE, TFT_BLACK);
		tft.setTextDatum(TL_DATUM);
		if(gps.location.isValid()){
			tft.fillRect(0, 0, 135, 120, TFT_GREEN);
		}
		else{
			tft.fillRect(0, 0, 135, 120, TFT_RED);
		}
		
		tft.drawString("Fix Status:", 2, 2);
		tft.setTextColor(gps.location.isValid() ? TFT_GREEN : TFT_RED, TFT_BLACK);
		tft.drawString(gps.location.isValid() ? "Valid" : "Invalid", 2, 12);
		tft.setTextColor(TFT_YELLOW, TFT_BLACK);
		tft.drawString("Lat: " + String(gps.location.lat(), 6), 2, 22);
		tft.drawString("Lng: " + String(gps.location.lng(), 6), 2, 32);
		tft.drawString("Alt: " + String(gps.altitude.meters(), 2) + "m", 2, 42);
		tft.drawString("Sats: " + String(gps.satellites.value()), 2, 52);
		tft.drawString("TimeStamp: " + String(gps.date.year()) + "-" + String(gps.date.month()) + "-" + String(gps.date.day()) + " " + String(gps.time.hour()) + ":" + String(gps.time.minute()) + ":" + String(gps.time.second()), 2, 62);
		


		if(roll < 30 && roll > -30 && pitch < 30 && pitch > -30 && flag == false){
			tft.fillRect(0, 120, 135, 120, TFT_GREEN);
		}
		else if(flag == true){
			tft.fillRect(0, 120, 135, 120, TFT_ORANGE);
		}
		else{
			tft.fillRect(0, 120, 135, 120, TFT_RED);
			flag = true;
		}
		tft.setTextColor(TFT_WHITE, TFT_BLACK);
		tft.setTextDatum(TL_DATUM);
		tft.drawString("Orientation:", 2, 122);
		tft.setTextColor(TFT_YELLOW, TFT_BLACK);
		tft.drawString("Pitch: " + String(pitch, 2), 2, 132);
		tft.drawString("Roll: " + String(roll, 2), 2, 142);
		tft.drawString("Yaw: " + String(yaw, 2), 2, 152);
		// Sample Data Every 1000ms for 5mins, and write to buffer

		data_buffer[buff_index].lat = gps.location.lat();
		data_buffer[buff_index].lng = gps.location.lng();
		data_buffer[buff_index].alt = gps.altitude.meters();
		data_buffer[buff_index].sats = gps.satellites.value();
		data_buffer[buff_index].timestamp = String(gps.date.year()) + "-" + String(gps.date.month()) + "-" + String(gps.date.day()) + " " + String(gps.time.hour()) + ":" + String(gps.time.minute()) + ":" + String(gps.time.second());

		buff_index++;		
	}

	if(buff_index == 5){
		// Try sending data to server else store in flash
		int size = sizeof(data_buffer);
		Serial.println(size);
		http.begin("http://94.100.26.221:8001/api/append");
		http.setTimeout(100000);
		http.addHeader("Content-Type", "application/x-www-form-urlencoded");
		String payload = "[";
		for(int i = 0; i < 6; i++){
			payload += "{";
			payload += "\"deviceID\": \"" + deviceID + "\",";
			payload += "\"lat\": \"" + String(data_buffer[i].lat, 6) + "\",";
			payload += "\"lon\": \"" + String(data_buffer[i].lng, 6) + "\",";
			payload += "\"intTemp\": \"" + String(data_buffer[i].internal_temp, 2) + "\",";
			payload += "\"extTemp\": \"" + String(data_buffer[i].external_temp, 2) + "\",";
			payload += "\"hum\": \"" + String(data_buffer[i].hum, 2) + "\",";
			payload += "\"maxXAccl\": \"" + String(data_buffer[i].max_ax, 2) + "\",";
			payload += "\"maxYAccl\": \"" + String(data_buffer[i].max_ay, 2) + "\",";
			payload += "\"maxZAccl\": \"" + String(data_buffer[i].max_az, 2) + "\",";
			payload += "\"pitch\": \"" + String(data_buffer[i].max_pitch, 2) + "\",";
			payload += "\"roll\": \"" + String(data_buffer[i].max_roll, 2) + "\",";
			payload += "\"yaw\": \"" + String(data_buffer[i].max_yaw, 2) + "\",";
			payload += "\"alt\": \"" + String(data_buffer[i].alt, 2) + "\",";
			payload += "\"satellites\": \"" + String(data_buffer[i].sats) + "\",";
			payload += "\"timestamp\": \"" + data_buffer[i].timestamp + "\",";
			payload += "\"capabilities\": \"GPS, Temperature, Accelerometer, Gyroscope\"";

			payload += "}";
			if(i != 5){
				payload += ",";
			}
		}
		payload += "]";
		Serial.println(payload);
		//String POSTPayload = "{\"device_id\": \"" + deviceID + "\", \"password\": \"" + password + "\", \"logs\": " + payload + "}";
		String POSTPayload = "deviceID=" + deviceID + "&password=" + password + "&logs=" + payload;
		int httpResponseCode = http.POST(POSTPayload);
		if (httpResponseCode > 199) {
			Serial.print("HTTP Response code: ");
			Serial.println(httpResponseCode);
			String response = http.getString();
			Serial.println(response);
		} else {
			Serial.print("Error code: ");
			Serial.println(httpResponseCode);
		}
		http.end();
		// If data is sent successfully, clear buffer
		if (httpResponseCode == 200) {
			buff_index = 0;
		}
		// If data is not sent successfully, store in flash
		else {
			buff_index = 0;
			// Write to LittleFS that device is initialized
			File file = LittleFS.open("/data.txt", "rw+");
			if (!file) {
				Serial.println("Failed to open file for writing");
				return;
			}

			// Read existing payload from file
			String existingPayload = file.readString();
			if (existingPayload != "") {
				payload.remove(0, 1);
				existingPayload.remove(existingPayload.length() - 1, 1);
				existingPayload += ",";
			} 			
			// Merge existing payload with new payload
			String mergedPayload = existingPayload + payload;

			// Write merged payload to file
			file.seek(0);
			if (file.print(mergedPayload)) {
				Serial.println("Payload merged and written to file");
			} else {
				Serial.println("Write failed");
			}

			file.close();
		}
	}

	// Try sending data to server that was stored in flash
	file = LittleFS.open("/data.txt", "r+");
	file.seek(0);
	if (!file) {
		Serial.println("Failed to open file for reading");
		Serial.println("No data to send");
	}
	else{
		String line = file.readStringUntil('\n');
		if (line == "") {
			Serial.println("No data to send");
			return;
		}
		Serial.println(line);
		http.begin("http://94.100.26.221:8001/api/append");
		http.setTimeout(100000);
		http.addHeader("Content-Type", "application/x-www-form-urlencoded");
		//String PostPayload = "{\"device_id\": \"" + deviceID + "\", \"password\": \"" + password + "\", \"logs\": " + line + "}";
		String PostPayload = "deviceID=" + deviceID + "&password=" + password + "&logs=" + line;
		int httpResponseCode = http.POST(PostPayload);
		if (httpResponseCode > 200) {
			Serial.print("HTTP Response code: ");
			Serial.println(httpResponseCode);
			String response = http.getString();
			Serial.println(response);
		} else {
			Serial.print("Error code: ");
			Serial.println(httpResponseCode);
		}
		http.end();
		// If data is sent successfully, clear buffer
		if (httpResponseCode == 200) {
			clean = true;
		}
	}
	file.close();
	if (clean) {
		LittleFS.remove("/data.txt");
		clean = false;
	}
}