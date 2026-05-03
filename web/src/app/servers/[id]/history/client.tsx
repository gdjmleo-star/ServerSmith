"use client";
import { useState } from "react";
import { useParams } from "next/navigation";
import { TrafficChart } from "@/components/TrafficChart";
import { useHistory } from "@/hooks/useHistory";
import { DEFAULT_HISTORY_DAYS } from "@/lib/config";

const DAY_OPTIONS = [7, 14, 30, 60, 90];

export default function HistoryPage() {
  const id = Number(useParams()?.id);
  const [days, setDays] = useState(DEFAULT_HISTORY_DAYS);
  const { data, loading, error } = useHistory(id, days);
  return <TrafficChart data={data} loading={loading} error={error} days={days} onDaysChange={setDays} dayOptions={DAY_OPTIONS} />;
}
