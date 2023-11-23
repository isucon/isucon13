<?php

declare(strict_types=1);

namespace IsuPipe\Reaction;

use TypeError;
use UnexpectedValueException;

class ReactionModel
{
    public function __construct(
        public ?int $id = null,
        public ?string $emojiName = null,
        public ?int $userId = null,
        public ?int $livestreamId = null,
        public ?int $createdAt = null,
    ) {
    }

    /**
     * @param array<string, mixed> $row
     * @throws UnexpectedValueException
     */
    public static function fromRow(array $row): ReactionModel
    {
        try {
            return new ReactionModel(
                id: $row['id'] ?? null,
                emojiName: $row['emoji_name'] ?? null,
                userId: $row['user_id'] ?? null,
                livestreamId: $row['livestream_id'] ?? null,
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
