<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use JsonSerializable;
use TypeError;
use UnexpectedValueException;

class NGWord implements JsonSerializable
{
    public function __construct(
        public ?int $id = null,
        public ?int $userId = null,
        public ?int $livestreamId = null,
        public ?string $word = null,
        public ?int $createdAt = null,
    ) {
    }

    /**
     * @param array<string, mixed> $row
     * @throws UnexpectedValueException
     */
    public static function fromRow(array $row): NGWord
    {
        try {
            return new NGWord(
                id: $row['id'] ?? null,
                userId: $row['user_id'] ?? null,
                livestreamId: $row['livestream_id'] ?? null,
                word: $row['word'] ?? null,
                createdAt: $row['created_at'] ?? null,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'id' => $this->id,
            'user_id' => $this->userId,
            'livestream_id' => $this->livestreamId,
            'word' => $this->word,
            'created_at' => $this->createdAt,
        ];
    }
}
