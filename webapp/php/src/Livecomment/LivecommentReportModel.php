<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use TypeError;
use UnexpectedValueException;

class LivecommentReportModel
{
    public function __construct(
        public ?int $id = null,
        public ?int $userId = null,
        public ?int $livestreamId = null,
        public ?int $livecommentId = null,
        public ?int $createdAt = null,
    ) {
    }

    /**
     * @param array<string, mixed> $row
     * @throws UnexpectedValueException
     */
    public static function fromRow(array $row): LivecommentReportModel
    {
        try {
            return new LivecommentReportModel(
                id: $row['id'] ?? null,
                userId: $row['user_id'] ?? null,
                livestreamId: $row['livestream_id'] ?? null,
                livecommentId: $row['livecomment_id'] ?? null,
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
