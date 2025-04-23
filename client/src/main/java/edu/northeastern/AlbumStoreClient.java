package edu.northeastern;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.mime.MultipartEntityBuilder;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.impl.conn.PoolingHttpClientConnectionManager;
import org.apache.http.util.EntityUtils;

import java.io.InputStream;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * Client for interacting with the Album Store API.
 * This class handles all HTTP communications with the server,
 * including connection pooling and request execution.
 */
public class AlbumStoreClient implements AutoCloseable {
  private static final String IMAGE_FILENAME = "Example.jpg";

  private final String baseUrl;
  private final CloseableHttpClient httpClient;
  private final byte[] imageBytes;
  private final Gson gson = new Gson();

  /**
   * Creates a new AlbumStoreClient with optimized connection pooling based on the expected load.
   *
   * @param baseUrl The base URL of the Album Store API
   * @param numRequestsPerThread Number of requests per thread
   */
  public AlbumStoreClient(String baseUrl,
                          int numRequestsPerThread) {
    if (baseUrl == null || baseUrl.isEmpty()) {
      throw new IllegalArgumentException("Base URL cannot be null or empty");
    }
    this.baseUrl = baseUrl;

    // Load the test image file from resources
    // This image will be reused for all POST requests to optimize memory usage
    try (InputStream imageStream = getClass().getClassLoader().getResourceAsStream(IMAGE_FILENAME)) {
      if (imageStream == null) {
        throw new IllegalStateException("Image file not found");
      }

      this.imageBytes = imageStream.readAllBytes();
    } catch (Exception e) {
      throw new RuntimeException(e);
    }

    // Configure connection pooling
    PoolingHttpClientConnectionManager connectionManager = new PoolingHttpClientConnectionManager();
    connectionManager.setMaxTotal(numRequestsPerThread);
    connectionManager.setDefaultMaxPerRoute(numRequestsPerThread);

    // Build HTTP client with configuration
    this.httpClient = HttpClients.custom()
        .setConnectionManager(connectionManager)
        .build();
  }

  /**
   * Executes an asynchronous POST request to create a new album with image.
   * This method prepares a multipart/form-data request with album metadata
   * and image content.
   *
   * @return CompletableFuture containing metrics about the request execution
   *         including start time, request type, latency, and response status
   */
  public CompletableFuture<RequestMetrics> createAlbumAsync() {
    return CompletableFuture.supplyAsync(() -> {
      HttpPost post = new HttpPost(baseUrl + "/albums");

      Map<String, Object> album = new HashMap<>();
      album.put("artist_id", "00000000-0000-0000-0000-000000000003");
      album.put("title", "Test Album");
      album.put("year", 2025);
      String albumJson = gson.toJson(album);

      MultipartEntityBuilder builder = MultipartEntityBuilder.create();
      builder.addBinaryBody(
          "image", imageBytes,
          ContentType.MULTIPART_FORM_DATA, IMAGE_FILENAME
      );
      builder.addTextBody("album", albumJson,
          ContentType.APPLICATION_JSON);
      post.setEntity(builder.build());

      // Record start time just before executing the request
      long startTime = System.currentTimeMillis();
      try (CloseableHttpResponse response = httpClient.execute(post)) {
        long endTime = System.currentTimeMillis();
        int statusCode = response.getStatusLine().getStatusCode();

        if (statusCode == 200 || statusCode == 201) {
          String responseBody = EntityUtils.toString(response.getEntity());
          try {
            JsonObject jsonResponse = gson.fromJson(responseBody, JsonObject.class);
            if (jsonResponse.has("id")) {
              String albumId = jsonResponse.get("id").getAsString();
              AlbumIdTracker.addAlbumId(albumId);
            }
          } catch (Exception e) {
            System.err.println("Failed to parse album ID: " + e.getMessage());
          }
        } else {
          // Ensure the response is fully consumed to release the connection
          EntityUtils.consume(response.getEntity());
        }

        return new RequestMetrics(
            startTime, "ALBUM_POST", endTime - startTime, statusCode
        );
      } catch (Exception e) {
        System.err.println("ALBUM_POST request failed: " + e.getMessage());
        return new RequestMetrics(
            startTime, "ALBUM_POST", -1, 500
        );
      }
    });
  }

  /**
   * Executes an asynchronous GET request to retrieve album information.
   * Fetches details for a specific album from the server.
   *
   * @return CompletableFuture containing metrics about the request execution
   *         including start time, request type, latency, and response status
   */
  public CompletableFuture<RequestMetrics> getAlbumAsync() {
    return CompletableFuture.supplyAsync(() -> {
      String albumId = AlbumIdTracker.getRandomAlbumId();
      HttpGet get = new HttpGet(baseUrl + "/albums/" + albumId);

      // Record start time just before executing the request
      long startTime = System.currentTimeMillis();
      try (CloseableHttpResponse response = httpClient.execute(get)) {
        long endTime = System.currentTimeMillis();
        int statusCode = response.getStatusLine().getStatusCode();

        // Ensure the response is fully consumed to release the connection
        EntityUtils.consume(response.getEntity());

        return new RequestMetrics(
            startTime, "ALBUM_GET", endTime - startTime, statusCode
        );
      } catch (Exception e) {
        System.err.println("GET request failed: " + e.getMessage());
        return new RequestMetrics(
            startTime, "ALBUM_GET", -1, 500
        );
      }
    });
  }

  /**
   * Executes an asynchronous POST request to review an album.
   * This method allows the user to like or dislike an album.
   *
   * @param reviewType The type of review ("like" or "dislike")
   * @return CompletableFuture containing metrics about the request execution
   *         including start time, request type, latency, and response status
   * @throws IllegalArgumentException If reviewType is not "like" or "dislike"
   */
  public CompletableFuture<RequestMetrics> reviewAlbumAsync(String reviewType) {
    return CompletableFuture.supplyAsync(() -> {
      if (!reviewType.equals("like") && !reviewType.equals("dislike")) {
        throw new IllegalArgumentException("Review type must be 'like' or 'dislike'");
      }

      String albumId = AlbumIdTracker.getRandomAlbumId();
      HttpPost post = new HttpPost(baseUrl + "/albums/" + albumId + "/" + reviewType);

      // Record start time just before executing the request
      long startTime = System.currentTimeMillis();
      try (CloseableHttpResponse response = httpClient.execute(post)) {
        long endTime = System.currentTimeMillis();
        int statusCode = response.getStatusLine().getStatusCode();

        // Ensure the response is fully consumed to release the connection
        EntityUtils.consume(response.getEntity());

        return new RequestMetrics(
            startTime, "REVIEW_POST", endTime - startTime, statusCode
        );
      } catch (Exception e) {
        System.err.println("REVIEW_POST request failed for album " + albumId + ":" + e.getMessage());
        return new RequestMetrics(
            startTime, "REVIEW_POST", -1, 500
        );
      }
    });
  }

  /**
   * Executes an asynchronous GET request to retrieve like/dislike counts for an album.
   * This method queries the review statistics for a given album ID.
   *
   * @return CompletableFuture containing metrics about the request execution
   *         including start time, request type, latency, and response status
   */
  public RequestMetrics getAlbumReviews() {
    String albumId = AlbumIdTracker.getRandomAlbumId();
    HttpGet get = new HttpGet(baseUrl + "/albums/" + albumId + "/review");

    // Record start time just before executing the request
    long startTime = System.currentTimeMillis();
    try (CloseableHttpResponse response = httpClient.execute(get)) {
      long endTime = System.currentTimeMillis();
      int statusCode = response.getStatusLine().getStatusCode();

      // Ensure the response is fully consumed to release the connection
      EntityUtils.consume(response.getEntity());

      // For debugging purposes if needed
//      String responseBody = EntityUtils.toString(response.getEntity());
//      System.out.println("REVIEW GET response: " + responseBody);

      return new RequestMetrics(
          startTime, "REVIEW_GET", endTime - startTime, statusCode
      );
    } catch (Exception e) {
      System.err.println("REVIEW GET request failed for album " + albumId + ": " + e.getMessage());
      return new RequestMetrics(
          startTime, "REVIEW_GET", -1, 500
      );
    }
  }

  /**
   * Closes the HTTP client and releases all system resources.
   * This method should be called when the client is no longer needed.
   */
  @Override
  public void close() throws Exception {
    if (httpClient != null) {
      httpClient.close();
    }
  }
}
