<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use TypeError;
use UnexpectedValueException;

class LivecommentModel
{
    public function __construct(
        public ?int $id = null,
        public ?int $userId = null,
        public ?int $livestreamId = null,
        public ?string $comment = null,
        public ?int $tip = null,
        public ?int $createdAt = null,
    ) {
    }

    /**
     * @param array<string, mixed> $row
     * @throws UnexpectedValueException
     */
    public static function fromRow(array $row): LivecommentModel
    {
        try {
            return new LivecommentModel(
                id: $row['id'] ?? null,
                userId: $row['user_id'] ?? null,
                livestreamId: $row['livestream_id'] ?? null,
                comment: $row['comment'] ?? null,
                tip: $row['tip'] ?? null,
                createdAt: $row['created_at'] ?? null,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
