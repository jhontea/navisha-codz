import React, { useState, useRef, useEffect } from "react";

interface LazyImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  placeholder?: string;
  rootMargin?: string;
}

export const LazyImage = React.memo(function LazyImage({
  placeholder,
  rootMargin = "100px",
  className = "",
  alt = "",
  src,
  ...props
}: LazyImageProps) {
  const [isLoaded, setIsLoaded] = useState(false);
  const [isInView, setIsInView] = useState(false);
  const imgRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = imgRef.current;
    if (!container) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsInView(true);
          observer.disconnect();
        }
      },
      { rootMargin }
    );

    observer.observe(container);
    return () => observer.disconnect();
  }, [rootMargin]);

  return (
    <div ref={imgRef} className={`relative overflow-hidden ${className}`}>
      {/* Placeholder / blur hash */}
      {!isLoaded && placeholder && (
        <div
          className="absolute inset-0 bg-cover bg-center filter blur-lg scale-110"
          style={{ backgroundImage: `url(${placeholder})` }}
        />
      )}
      {!isLoaded && !placeholder && (
        <div className="absolute inset-0 bg-slate-200 animate-pulse" />
      )}
      {isInView && (
        <img
          src={src}
          alt={alt}
          loading="lazy"
          decoding="async"
          onLoad={() => setIsLoaded(true)}
          className={`transition-opacity duration-300 ${isLoaded ? "opacity-100" : "opacity-0"}`}
          {...props}
        />
      )}
    </div>
  );
});
