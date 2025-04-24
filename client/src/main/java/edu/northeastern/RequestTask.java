package edu.northeastern;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CompletableFuture;

/**
 * Runnable task that executes a series of GET and POST requests.
 * Each task represents a single thread's workload in the performance test.
 */
public class RequestTask implements Runnable {
  private final AlbumStoreClient client;
  private final int requestCount;
  private final List<CompletableFuture<RequestMetrics>> futures;

  /**
   * Creates a new task to execute a specified number of request pairs.
   * Each pair consists of one POST request followed by one GET request.
   *
   * @param client The client to use for making HTTP requests
   * @param requestCount The number of GET/POST pairs to execute
   * @param futures The shared list to store metrics for all requests
   */
  public RequestTask(AlbumStoreClient client, int requestCount,
                     List<CompletableFuture<RequestMetrics>> futures) {
    if (client == null) {
      throw new IllegalArgumentException("client cannot be null");
    }
    if (requestCount < 1) {
      throw new IllegalArgumentException("requestCount cannot be less than 1");
    }
    if (futures == null) {
      throw new IllegalArgumentException("futures cannot be null");
    }

    this.client = client;
    this.requestCount = requestCount;
    this.futures = futures;
  }

  @Override
  public void run() {
    try {
      List<CompletableFuture<RequestMetrics>> threadFutures = new ArrayList<>(requestCount * 2);

      // Execute the specified number of request pairs
      for (int i = 0; i < requestCount; i++) {
        threadFutures.add(client.createAlbumAsync());
//        threadFutures.add(client.getAlbumAsync());
        threadFutures.add(client.reviewAlbumAsync("like"));
        threadFutures.add(client.reviewAlbumAsync("dislike"));
      }

      synchronized (futures) {
        futures.addAll(threadFutures);
      }
    } catch (Exception e) {
      System.err.println("Error in RequestTask: " + e.getMessage());
      throw new RuntimeException(e);
    }
  }
}
