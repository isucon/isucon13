<?php

declare(strict_types=1);

namespace IsuPipe\Stats;

/**
 * @phpstan-type UserRanking list<UserRankingEntry>
 */
class UserRankingEntry
{
    public function __construct(
        public string $username,
        public int $score,
    ) {
    }

    public function compare(UserRankingEntry $other): int
    {
        if ($this->score === $other->score) {
            return $this->username <=> $other->username;
        }

        return $this->score <=> $other->score;
    }
}
