<?php

declare(strict_types=1);

namespace IsuPipe\Payment;

use Jsonserializable;

class PaymentResult implements Jsonserializable
{
    public function __construct(
        public int $totalTip,
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'total_tip' => $this->totalTip,
        ];
    }
}
