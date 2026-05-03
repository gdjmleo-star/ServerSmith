import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface StatCardProps {
  title: string;
  value: string | number;
  color?: string;
}

export default function StatCard({ title, value, color }: StatCardProps) {
  return (
    <Card className="bg-zinc-950 border-zinc-800">
      <CardHeader className="p-3 pb-0">
        <CardTitle className="text-xs font-normal text-gray-500">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="p-3 pt-1">
        <p className={`text-xl font-semibold ${color || "text-white"}`}>
          {value}
        </p>
      </CardContent>
    </Card>
  );
}
