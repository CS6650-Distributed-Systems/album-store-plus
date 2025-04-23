package edu.northeastern;

/**
 * Represents metrics collected for a single HTTP request during performance testing.
 * This class captures essential timing and status information about each request,
 * which can be used for analyzing the system's performance characteristics including
 * latency, throughput, and error rates.
 */
public class RequestMetrics {
  private final long startTime;
  private final String requestType;
  private final long latency;
  private final long responseCode;

  /**
   * Creates a new RequestMetrics instance with the specified metrics.
   *
   * @param startTime Timestamp when the request was initiated (milliseconds since epoch)
   * @param requestType Type of the HTTP request (GET or POST)
   * @param latency Time taken to complete the request in milliseconds
   * @param responseCode HTTP response status code
   */
  public RequestMetrics(long startTime, String requestType, long latency, long responseCode) {
    this.startTime = startTime;
    this.requestType = requestType;
    this.latency = latency;
    this.responseCode = responseCode;
  }

  public long getStartTime() { return startTime; }
  public String getRequestType() { return requestType; }
  public long getLatency() { return latency; }
  public long getResponseCode() { return responseCode; }

  public boolean isSuccessful() {
    return responseCode == 200 || responseCode == 201;
  }

  @Override
  public String toString() {
    return String.format(
        "%d,%s,%d,%d",
        startTime, requestType, latency, responseCode
    );
  }
}
