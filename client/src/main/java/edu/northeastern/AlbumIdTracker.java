package edu.northeastern;

import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ThreadLocalRandom;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * A thread-safe utility class for tracking and retrieving album IDs.
 * This class maintains a collection of album IDs that have been created during testing
 * and provides methods to safely add new IDs and randomly select existing ones.
 * It's designed to support concurrent operations without explicit synchronization.
 */
public class AlbumIdTracker {
  private static final ConcurrentHashMap<Integer, Long> createdAlbumIds = new ConcurrentHashMap<>();
  private static final AtomicInteger counter = new AtomicInteger(0);

  /**
   * Adds a new album ID to the tracker.
   * Thread-safe implementation for concurrent additions.
   *
   * @param id The album ID to add to the collection
   */
  public static void addAlbumId(long id) {
    createdAlbumIds.put(counter.getAndIncrement(), id);
  }

  /**
   * Retrieves a random album ID from the collection of tracked IDs.
   * If no album IDs have been added yet, returns 0 as a default value.
   * Uses ThreadLocalRandom for better performance in concurrent environments.
   *
   * @return A randomly selected album ID, or 0 if none exists
   */
  public static long getRandomAlbumId() {
    if (createdAlbumIds.isEmpty()) {
      return 0;
    }

    int size = createdAlbumIds.size();
    int randomIndex = ThreadLocalRandom.current().nextInt(size);

    return createdAlbumIds.get(randomIndex);
  }

  /**
   * Returns the map containing all tracked album IDs.
   * This method is primarily intended for testing or status reporting.
   *
   * @return The concurrent map of all tracked album IDs
   */
  public static ConcurrentHashMap<Integer, Long> getCreatedAlbumIds() {
    return createdAlbumIds;
  }
}
