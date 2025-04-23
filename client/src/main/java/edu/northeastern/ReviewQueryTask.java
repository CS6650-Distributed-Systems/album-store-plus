package edu.northeastern;

import java.util.List;

/**
 * A task that continuously queries album review information using random valid album IDs.
 * This class is designed to be used as a separate thread that runs after the main album creation
 * phase of the performance test has completed. It continuously retrieves review statistics for
 * albums to measure the performance of GET review operations.
 */
public class ReviewQueryTask implements Runnable {
  private final AlbumStoreClient client;
  private final List<RequestMetrics> queryMetrics;

  /**
   * Creates a new review query task.
   *
   * @param client The client to use for making HTTP requests to the album review API
   * @param queryMetrics The shared list to store metrics for all review queries
   * @throws IllegalArgumentException If client or futures is null
   */
  public ReviewQueryTask(AlbumStoreClient client,
                         List<RequestMetrics> queryMetrics) {
    if (client == null) {
      throw new IllegalArgumentException("client cannot be null");
    }
    if (queryMetrics == null) {
      throw new IllegalArgumentException("queryMetrics cannot be null");
    }

    this.client = client;
    this.queryMetrics = queryMetrics;
  }

  @Override
  public void run() {
    try {
      while (!Thread.currentThread().isInterrupted()) {
        // Execute the review GET request asynchronously
        RequestMetrics queryMetric =
            client.getAlbumReviews();

        synchronized (queryMetrics) {
          queryMetrics.add(queryMetric);
        }
      }
    } catch (Exception e) {
      System.err.println("Error in ReviewQueryTask: " + e.getMessage());
    }
  }
}
