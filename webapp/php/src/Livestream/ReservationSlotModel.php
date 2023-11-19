<?php

declare(strict_types=1);

namespace IsuPipe\Livestream;

use TypeError;
use UnexpectedValueException;

class ReservationSlotModel
{
    public function __construct(
        public ?int $id = null,
        public ?int $slot = null,
        public ?int $startAt = null,
        public ?int $endAt = null,
    ) {
    }

    /**
     * @param array<string, mixed> $row
     * @throws UnexpectedValueException
     */
    public static function fromRow(array $row): ReservationSlotModel
    {
        try {
            return new ReservationSlotModel(
                id: $row['id'] ?? null,
                slot: $row['slot'] ?? null,
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
