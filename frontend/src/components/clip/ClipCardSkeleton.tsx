import { Skeleton } from '@/components/ui';

export function ClipCardSkeleton() {
  return (
    <div className="bg-card border border-border p-4">
      <div className="flex gap-4">
        {/* Vote sidebar */}
        <div className="flex flex-col items-center gap-2 w-10">
          <Skeleton variant="rectangular" width={32} height={32} className="rounded" />
          <Skeleton variant="text" width={32} height={20} />
          <Skeleton variant="rectangular" width={32} height={32} className="rounded" />
        </div>

        {/* Main content */}
        <div className="flex-1">
          {/* Thumbnail */}
          <Skeleton variant="rectangular" width="100%" height={200} className="rounded-lg mb-3" />

          {/* Title */}
          <Skeleton variant="text" width="80%" height={24} className="mb-2" />

          {/* Metadata */}
          <div className="flex gap-2 mb-3">
            <Skeleton variant="text" width={100} height={16} />
            <Skeleton variant="text" width={80} height={16} />
            <Skeleton variant="text" width={60} height={16} />
          </div>

          {/* Action bar */}
          <div className="flex gap-4">
            <Skeleton variant="text" width={80} height={20} />
            <Skeleton variant="text" width={80} height={20} />
            <Skeleton variant="text" width={60} height={20} />
          </div>
        </div>
      </div>
    </div>
  );
}
