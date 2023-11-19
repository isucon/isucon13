<?php

declare(strict_types=1);

namespace IsuPipe\Top;

use TypeError;
use UnexpectedValueException;

class TagModel
{
    public function __construct(
        public ?int $id = null,
        public ?string $name = null,
    ) {
    }

    /**
     * @param array<string, mixed> $row
     * @throws UnexpectedValueException
     */
    public static function fromRow(array $row): TagModel
    {
        try {
            return new TagModel(
                id: $row['id'] ?? null,
                name: $row['name'] ?? null,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
