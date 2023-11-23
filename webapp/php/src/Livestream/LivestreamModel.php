<?php

declare(strict_types=1);

namespace IsuPipe\Livestream;

use TypeError;
use UnexpectedValueException;

class LivestreamModel
{
    public function __construct(
        public ?int $id = null,
        public ?int $userId = null,
        public ?string $title = null,
        public ?string $description = null,
        public ?string $playlistUrl = null,
        public ?string $thumbnailUrl = null,
        public ?int $startAt = null,
        public ?int $endAt = null,
    ) {
    }

    /**
     * @param array<string, mixed> $row
     * @throws UnexpectedValueException
     */
    public static function fromRow(array $row): LivestreamModel
    {
        try {
            return new LivestreamModel(
                id: $row['id'] ?? null,
                userId: $row['user_id'] ?? null,
                title: $row['title'] ?? null,
                description: $row['description'] ?? null,
                playlistUrl: $row['playlist_url'] ?? null,
                thumbnailUrl: $row['thumbnail_url'] ?? null,
                startAt: $row['start_at'] ?? null,
                endAt: $row['end_at'] ?? null,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
