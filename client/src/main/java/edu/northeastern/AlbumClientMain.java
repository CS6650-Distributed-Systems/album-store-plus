package edu.northeastern;

public class AlbumClientMain {
  public static void main(String[] args) {

    int threadGroupSize = 10;
    int numThreadGroups = 10;
    int delay = 2;
    String ipAddr = "localhost";

    if (args.length == 4) {
      threadGroupSize = Integer.parseInt(args[0]);
      numThreadGroups = Integer.parseInt(args[1]);
      delay = Integer.parseInt(args[2]);
      ipAddr = args[3];
//    } else {
//      System.err.println("Usage: java AlbumClientMain <threadGroupSize> <numThreadGroups> <delay> <ipAddr>");
//      System.exit(1);
    }

      PerformanceTest test = new PerformanceTest(threadGroupSize, numThreadGroups, delay, ipAddr);
      test.runTest();
  }
}