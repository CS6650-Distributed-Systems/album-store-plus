const { S3Client, GetObjectCommand, PutObjectCommand } = require('@aws-sdk/client-storage');
const { v4: uuidv4 } = require('uuid');
const sharp = require('sharp');
const stream = require('stream');

// Configuration values matching the Go implementation
const maxWidth = 100;
const maxHeight = 100;
const quality = 85;

// Create S3 client
const s3Client = new S3Client();

// Helper function to stream S3 data through Sharp
async function processImage(bucketName, sourceKey, destinationKey) {
    try {
        // Download the original image
        const getObjectParams = {
            Bucket: bucketName,
            Key: sourceKey
        };

        const { Body } = await s3Client.send(new GetObjectCommand(getObjectParams));

        // Convert the readable stream to a buffer
        const chunks = [];
        for await (const chunk of Body) {
            chunks.push(chunk);
        }
        const buffer = Buffer.concat(chunks);

        // Process the image with Sharp (equivalent to the Go version)
        const processedImageBuffer = await sharp(buffer)
            .resize({
                width: maxWidth,
                height: maxHeight,
                fit: 'inside',
                withoutEnlargement: true
            })
            .jpeg({ quality: quality })
            .toBuffer();

        // Upload the processed image
        const putObjectParams = {
            Bucket: bucketName,
            Key: destinationKey,
            Body: processedImageBuffer,
            ContentType: 'image/jpeg'
        };

        await s3Client.send(new PutObjectCommand(putObjectParams));
        return true;
    } catch (error) {
        console.error('Error processing image:', error);
        throw error;
    }
}

// Lambda handler
exports.handler = async (event) => {
    try {
        // Generate a unique process ID
        const processId = uuidv4();

        // Extract request parameters
        const { sourceKey, destinationKey } = event;

        // Get bucket name from environment variable
        const bucketName = process.env.S3_BUCKET_NAME;
        if (!bucketName) {
            throw new Error('S3_BUCKET_NAME environment variable not set');
        }

        // Process the image
        await processImage(bucketName, sourceKey, destinationKey);

        // Return success response
        return {
            processId: processId,
            status: 'completed'
        };
    } catch (error) {
        console.error('Lambda execution error:', error);

        // Return error response
        return {
            processId: uuidv4(),
            status: 'error',
            error: error.message
        };
    }
};
