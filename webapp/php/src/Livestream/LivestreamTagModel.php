<?php

declare(strict_types=1);

namespace IsuPipe\Livestream;

use TypeError;
use UnexpectedValueException;

class LivestreamTagModel
{
    public function __construct(
        public ?int $id = null,
        public ?int $livestreamId = null,
        public ?int $tagId = null,
    ) {
    }

    /**
     * @param array<string, mixed> $row
     * @throws UnexpectedValueException
     */
    public static function fromRow(array $row): LivestreamTagModel
    {
        try {
            return new LivestreamTagModel(
                id: $row['id'] ?? null,
                livestreamId: $row['livestream_id'] ?? null,
                tagId: $row['tag_id'] ?? null,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
