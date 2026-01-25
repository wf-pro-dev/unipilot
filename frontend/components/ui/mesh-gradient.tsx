// components/ui/mesh-background.tsx
'use client'

import { useState, useEffect } from 'react';

interface MeshBackgroundProps {
  animated?: boolean;
  interactive?: boolean;
  density?: 'sparse' | 'balanced' | 'dense';
  showOrbs?: boolean;
  showGrid?: boolean;
}

export function MeshBackground({
  animated = true,
  interactive = false,
  density = 'balanced',
  showOrbs = true,
  showGrid = true
}: MeshBackgroundProps) {
  const [mousePosition, setMousePosition] = useState({ x: 0, y: 0 });
  
  // Interactive mouse tracking
  useEffect(() => {
    if (!interactive) return;
    
    const handleMouseMove = (e: MouseEvent) => {
      setMousePosition({
        x: e.clientX,
        y: e.clientY
      });
    };
    
    window.addEventListener('mousemove', handleMouseMove);
    return () => window.removeEventListener('mousemove', handleMouseMove);
  }, [interactive]);
  
  // Get grid size based on density
  const gridSize = {
    sparse: 120,
    balanced: 80,
    dense: 40
  }[density];
  
  // Dynamic styles for the mesh pattern
  const meshStyle = {
    '--mesh-grid-size': `${gridSize}px`
  } as React.CSSProperties;
  
  return (
    <div 
      className={` mesh-background ${animated ? 'animated' : ''} ${interactive ? 'interactive' : ''}`}
      style={meshStyle}
    >
      {/* Animated orbs */}
      {showOrbs && (
        <>
          <div className="mesh-orb-1" />
          <div className="mesh-orb-2" />
          <div className="mesh-orb-3" />
        </>
      )}
      
      {/* Grid overlay */}
      {showGrid && <div className="mesh-grid" />}
      
      {/* Interactive mouse highlight */}
      {interactive && (
        <div 
          className="mouse-highlight"
          style={{
            left: `${mousePosition.x - 150}px`,
            top: `${mousePosition.y - 150}px`,
          }}
        />
      )}
      
 
    </div>
  );
}

