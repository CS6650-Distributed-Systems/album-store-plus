package edu.northeastern;

import java.io.BufferedWriter;
import java.io.FileWriter;
import java.io.IOException;
import java.util.List;

/**
 * Handles the persistent storage of request metrics to a CSV file.
 * Creates timestamped log files for each test run for later analysis.
 */
public class RequestLogger {
  private final String filename;

  /**
   * Creates a new logger instance with a unique, timestamped filename.
   * The filename format is: performance_test_[testName]_[timestamp].csv
   *
   * @param testName Identifier for this test run
   */
  public RequestLogger(String testName) {
    long timestamp = System.currentTimeMillis();
    this.filename = String.format("performance_test_%s_%s.csv", testName, timestamp);
  }

  /**
   * Writes a collection of request metrics to the log file in CSV format.
   * The CSV includes headers and one row per request with all collected metrics.
   *
   * @param metrics List of metrics to write to the file
   */
  public void writeMetricsToFile(List<RequestMetrics> metrics) {
    try (BufferedWriter writer = new BufferedWriter(new FileWriter(filename))) {
      writer.write("start_time,request_type,latency_ms,response_code\n");


      for (RequestMetrics metric : metrics) {
        writer.write(String.format("%d,%s,%d,%d\n",
            metric.getStartTime(),
            metric.getRequestType(),
            metric.getLatency(),
            metric.getResponseCode()
        ));
      }
    } catch (IOException e) {
      System.err.println("Error writing to log file: " + e.getMessage());
    }
  }
}
