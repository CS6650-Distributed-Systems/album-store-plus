package edu.northeastern;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.concurrent.*;
import java.util.stream.Collectors;

/**
 * Coordinates and manages the entire performance testing process.
 * This class handles both the warmup phase and the main testing phase,
 * collecting and analyzing performance metrics throughout the test execution.
 */
public class PerformanceTest {
  private static final int WARMUP_THREADS = 10;
  private static final int QUERY_REVIEW_THREADS = 3;
  private static final int WARMUP_REQUESTS = 100;
  private static final int TEST_REQUESTS = 100;

  // Configuration for the performance test
  private final int threadGroupSize;
  private final int numThreadGroups;
  private final int delay;
  private final String baseUrl;

  // Lists to store metrics for different phases
  private final List<CompletableFuture<RequestMetrics>> warmupMetrics;
  private final List<CompletableFuture<RequestMetrics>> runMetrics;
  private final List<RequestMetrics> queryMetrics;

  /**
   * Creates a new performance test with the specified configuration.
   *
   * @param threadGroupSize Number of threads in each group
   * @param numThreadGroups Number of thread groups to create
   * @param delay           lay between starting each thread grou
   * @param baseUrl         The base URL of the server to test
   */
  public PerformanceTest(int threadGroupSize, int numThreadGroups, int delay, String baseUrl) {
    this.threadGroupSize = threadGroupSize;
    this.numThreadGroups = numThreadGroups;
    this.delay = delay;
    this.baseUrl = baseUrl;

    this.warmupMetrics = Collections.synchronizedList(new ArrayList<>());
    this.runMetrics = Collections.synchronizedList(new ArrayList<>());
    this.queryMetrics = Collections.synchronizedList(new ArrayList<>());
  }

  public void runTest() {
    System.out.println("Starting performance test...");
    System.out.println("Beginning warmup phase...");
    warmup();
    System.out.println("Warmup phase completed.");

    loadTest();
  }

  /**
   * Executes the main load test phase of the performance test
   */
  private void loadTest() {
    RequestLogger logger = new RequestLogger(String.format("%d_%d", threadGroupSize, numThreadGroups));

    try (AlbumStoreClient client = new AlbumStoreClient(
        baseUrl, TEST_REQUESTS)) {
      ExecutorService executor = Executors.newFixedThreadPool(
          threadGroupSize * numThreadGroups);
      List<List<CompletableFuture<?>>> threadGroupFutures = new ArrayList<>();

      ExecutorService queryReviewExecutor = Executors.newFixedThreadPool(QUERY_REVIEW_THREADS);

      long startTime = System.currentTimeMillis();
      for (int group = 0; group < numThreadGroups; group++) {
        List<CompletableFuture<?>> groupFutures = new ArrayList<>();
        threadGroupFutures.add(groupFutures);
        for (int thread = 0; thread < threadGroupSize; thread++) {
          groupFutures.add(CompletableFuture.runAsync(
              new RequestTask(client, TEST_REQUESTS, runMetrics), executor));
        }

        if (group < numThreadGroups - 1) {
          TimeUnit.SECONDS.sleep(delay);
        }
      }

      // Wait for the first group to finish before starting query reviews
      CompletableFuture.allOf(
          threadGroupFutures.get(0).toArray(new CompletableFuture[0])
      ).join();

      System.out.println("First thread group completed, starting query threads...");

      // Start query review threads
      long queryReviewStartTime = System.currentTimeMillis();
      for (int i = 0; i < QUERY_REVIEW_THREADS; i++) {
        queryReviewExecutor.submit(new ReviewQueryTask(client, queryMetrics));
      }

      // Wait for all futures to complete
      for (List<CompletableFuture<?>> groupFutures : threadGroupFutures) {
        CompletableFuture.allOf(groupFutures.toArray(new CompletableFuture[0])).join();
      }
      CompletableFuture.allOf(runMetrics.toArray(new CompletableFuture[0])).join();
      executor.shutdown();

      // Wait for query review threads to finish
      queryReviewExecutor.shutdownNow();

      long endTime = System.currentTimeMillis();
      long wallTime = endTime - startTime;
      long queryReviewWallTime = endTime - queryReviewStartTime;

      List<RequestMetrics> completedMetrics = runMetrics.stream()
          .map(CompletableFuture::join)
          .collect(Collectors.toList());

      logger.writeMetricsToFile(completedMetrics);

      // Calculate total expected requests
      int totalRequests = threadGroupSize * numThreadGroups * TEST_REQUESTS * 4;
      long queryReviewTotalRequests = queryMetrics.size();

      displayResults(wallTime, totalRequests, runMetrics);
      displayQueryReviewResults(queryReviewWallTime, queryReviewTotalRequests, queryMetrics);
    } catch (Exception e) {
      System.err.println("Error during test execution: " + e.getMessage());
      throw new RuntimeException(e);
    }
  }

  /**
   * Executes the warmup phase of the performance test
   */
  private void warmup() {
    try (AlbumStoreClient client = new AlbumStoreClient(
        baseUrl, WARMUP_REQUESTS)) {
      ExecutorService executor = Executors.newFixedThreadPool(WARMUP_THREADS);
      List<CompletableFuture<?>> taskFutures = new ArrayList<>();
      for (int i = 0; i < WARMUP_THREADS; i++) {
        taskFutures.add(CompletableFuture.runAsync(
            new RequestTask(client, WARMUP_REQUESTS, warmupMetrics), executor));
      }

      // Wait for all futures to complete
      CompletableFuture.allOf(taskFutures.toArray(new CompletableFuture[0])).join();
      CompletableFuture.allOf(warmupMetrics.toArray(new CompletableFuture[0])).join();
      executor.shutdown();
    } catch (Exception e) {
      System.err.println("Error during warmup: " + e.getMessage());
      throw new RuntimeException(e);
    }
  }

  /**
   * Analyzes and displays comprehensive test results.
   * Calculates and presents:
   * 1. Basic Metrics:
   * - Total wall time
   * - Total request count
   * - Success/failure counts
   * - Success rate
   * - Overall throughput
   * 2. Detailed Analysis (per request type):
   * - Mean response time
   * - Median response time
   * - 99th percentile (P99)
   * - Min/Max response times
   *
   * @param wallTime Total execution time of the test in milliseconds
   */
  private void displayResults(long wallTime, long numRequests, List<CompletableFuture<RequestMetrics>> metrics) {
    // Calculate total successful requests
    long successfulRequests = metrics.stream()
        .map(CompletableFuture::join)
        .filter(RequestMetrics::isSuccessful)
        .count();

    // Calculate failed requests and success rate
    long failureCount = numRequests - successfulRequests;
    double successRate = (numRequests > 0) ?
        (double) successfulRequests / numRequests * 100 : 0.0;

    // Calculate throughput (requests per second)
    double throughput = (double) successfulRequests / (wallTime / 1000.0);

    // Get all successful request metrics for statistical analysis
    List<RequestMetrics> allMetrics = metrics.stream()
        .map(CompletableFuture::join)
        .filter(RequestMetrics::isSuccessful)
        .collect(Collectors.toList());

    // Group metrics by request type (GET/POST)
    Map<String, List<RequestMetrics>> metricsByType = allMetrics.stream()
        .collect(Collectors.groupingBy(RequestMetrics::getRequestType));

    // Display all performance metrics
    System.out.println("\nPerformance Test Results");
    System.out.println("=======================");

    // Display basic metrics
    System.out.println("\nBasic Metrics:");
    System.out.printf("Wall Time: %d ms%n", wallTime);
    System.out.printf("Total Requests: %d%n", numRequests);
    System.out.printf("Successful Requests: %d%n", successfulRequests);
    System.out.printf("Failed Requests: %d%n", failureCount);
    System.out.printf("Success Rate: %.2f%%%n", successRate);
    System.out.printf("Throughput: %.2f requests/second%n", throughput);

    // Display detailed statistical analysis
    System.out.println("\nDetailed Response Time Analysis:");
    for (Map.Entry<String, List<RequestMetrics>> entry : metricsByType.entrySet()) {
      String requestType = entry.getKey();
      List<Long> latencies = entry.getValue().stream()
          .map(RequestMetrics::getLatency)
          .sorted()
          .collect(Collectors.toList());

      double mean = calculateMean(latencies);
      double median = calculateMedian(latencies);
      double p99 = calculatePercentile(latencies, 0.99);
      long min = latencies.get(0);
      long max = latencies.get(latencies.size() - 1);

      // Display query review statistics
      System.out.printf("\n%s Request Statistics:\n", requestType);
      System.out.printf("Mean: %.2f ms\n", mean);
      System.out.printf("Median: %.2f ms\n", median);
      System.out.printf("p99: %.2f ms\n", p99);
      System.out.printf("Min: %d ms\n", min);
      System.out.printf("Max: %d ms\n", max);
    }
  }

  private void displayQueryReviewResults(long wallTime, long numRequests, List<RequestMetrics> metrics) {
    // Calculate total successful requests
    long successfulRequests = metrics.stream()
        .filter(RequestMetrics::isSuccessful)
        .count();

    // Calculate failed requests and success rate
    long failureCount = numRequests - successfulRequests;
    double successRate = (numRequests > 0) ?
        (double) successfulRequests / numRequests * 100 : 0.0;

    // Calculate throughput (requests per second)
    double throughput = (double) successfulRequests / (wallTime / 1000.0);

    // Display all performance metrics
    System.out.println("\nQuery Review Results");
    System.out.println("=======================");

    // Display basic metrics
    System.out.println("\nBasic Metrics:");
    System.out.printf("Wall Time: %d ms%n", wallTime);
    System.out.printf("Total Requests: %d%n", numRequests);
    System.out.printf("Successful Requests: %d%n", successfulRequests);
    System.out.printf("Failed Requests: %d%n", failureCount);
    System.out.printf("Success Rate: %.2f%%%n", successRate);
    System.out.printf("Throughput: %.2f requests/second%n", throughput);

    // Display detailed statistical analysis
    System.out.println("\nDetailed Response Time Analysis:");
    List<Long> latencies = metrics.stream()
        .filter(RequestMetrics::isSuccessful)
        .map(RequestMetrics::getLatency)
        .sorted()
        .collect(Collectors.toList());
    double mean = calculateMean(latencies);
    double median = calculateMedian(latencies);
    double p99 = calculatePercentile(latencies, 0.99);
    long min = latencies.get(0);
    long max = latencies.get(latencies.size() - 1);

    // Display query review statistics
    System.out.printf("\nQuery Review Request Statistics:\n");
    System.out.printf("Mean: %.2f ms\n", mean);
    System.out.printf("Median: %.2f ms\n", median);
    System.out.printf("p99: %.2f ms\n", p99);
    System.out.printf("Min: %d ms\n", min);
    System.out.printf("Max: %d ms\n", max);
  }

  /**
   * Calculates the mean value from a list of latencies
   *
   * @param values List of latency values
   * @return The mean value
   */
  private double calculateMean(List<Long> values) {
    return values.stream().mapToDouble(Long::doubleValue).average().orElse(0.0);
  }

  /**
   * Calculates the median value from a sorted list of latencies
   *
   * @param values Sorted list of latency values
   * @return The median value
   */
  private double calculateMedian(List<Long> values) {
    int size = values.size();
    if (size % 2 == 0) {
      return (values.get(size / 2 - 1) + values.get(size / 2)) / 2.0;
    } else {
      return values.get(size / 2);
    }
  }

  /**
   * Calculates the specified percentile from a sorted list of latencies
   *
   * @param values     Sorted list of latency values
   * @param percentile The percentile to calculate (e.g., 0.99 for P99)
   * @return The percentile value
   */
  private double calculatePercentile(List<Long> values, double percentile) {
    int index = (int) Math.ceil(percentile * values.size()) - 1;
    return values.get(index);
  }
}
