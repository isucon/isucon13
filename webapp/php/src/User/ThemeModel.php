<?php

declare(strict_types=1);

namespace IsuPipe\User;

use TypeError;
use UnexpectedValueException;

class ThemeModel
{
    public function __construct(
        public ?int $id = null,
        public ?int $userId = null,
        public ?bool $darkMode = null,
    ) {
    }

    /**
     * @param array<string, mixed> $row
     * @throws UnexpectedValueException
     */
    public static function fromRow(array $row): ThemeModel
    {
        try {
            return new ThemeModel(
                id: $row['id'] ?? null,
                userId: $row['user_id'] ?? null,
                darkMode: is_null($row['dark_mode']) ? null : (bool) $row['dark_mode'],
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
