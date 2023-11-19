<?php

declare(strict_types=1);

namespace IsuPipe\User;

use TypeError;
use UnexpectedValueException;

class UserModel
{
    public function __construct(
        public ?int $id = null,
        public ?string $name = null,
        public ?string $displayName = null,
        public ?string $description = null,
        public ?string $hashedPassword = null,
    ) {
    }

    /**
     * @param array<string, mixed> $row
     * @throws UnexpectedValueException
     */
    public static function fromRow(array $row): UserModel
    {
        try {
            return new UserModel(
                id: $row['id'] ?? null,
                name: $row['name'] ?? null,
                displayName: $row['display_name'] ?? null,
                description: $row['description'] ?? null,
                hashedPassword: $row['password'] ?? null,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
