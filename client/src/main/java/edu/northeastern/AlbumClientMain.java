package edu.northeastern;

public class AlbumClientMain {
  public static void main(String[] args) {
    if (args.length == 4) {
      int threadGroupSize = Integer.parseInt(args[0]);
      int numThreadGroups = Integer.parseInt(args[1]);
      int delay = Integer.parseInt(args[2]);
      String ipAddr = args[3];

      PerformanceTest test = new PerformanceTest(threadGroupSize, numThreadGroups, delay, ipAddr);
      test.runTest();
    } else {
      System.err.println("Usage: java AlbumClientMain <threadGroupSize> <numThreadGroups> <delay> <ipAddr>");
      System.exit(1);
    }
  }
}