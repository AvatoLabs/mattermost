// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useEffect, useRef, useState} from 'react';

import mermaid from 'mermaid';

type Props = {
    code: string;
    className?: string;
}

let mermaidInitialized = false;

const MermaidRenderer: React.FC<Props> = ({code, className = ''}: Props) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const [error, setError] = useState<string | null>(null);
    const [isRendering, setIsRendering] = useState(true);

    useEffect(() => {
        if (!mermaidInitialized) {
            mermaid.initialize({
                startOnLoad: false,
                theme: 'default',
                securityLevel: 'strict',
                fontFamily: 'inherit',
            });
            mermaidInitialized = true;
        }
    }, []);

    useEffect(() => {
        const renderDiagram = async () => {
            if (!containerRef.current || !code) {
                return;
            }

            setIsRendering(true);
            setError(null);

            try {
                // Generate a unique ID for this diagram
                const id = `mermaid-${Math.random().toString(36).substr(2, 9)}`;

                // Render the diagram using mermaid 9.4.3 API
                const svgCode = await mermaid.render(id, code);

                // Insert the SVG into the container
                if (containerRef.current) {
                    containerRef.current.innerHTML = svgCode;
                }
            } catch (err) {
                console.error('Mermaid rendering error:', err);
                setError(err instanceof Error ? err.message : 'Failed to render Mermaid diagram');
            } finally {
                setIsRendering(false);
            }
        };

        renderDiagram();
    }, [code]);

    if (error) {
        return (
            <div className={`mermaid-error ${className}`}>
                <div className='alert alert-warning'>
                    <strong>Mermaid Rendering Error:</strong>
                    <pre>{error}</pre>
                </div>
            </div>
        );
    }

    return (
        <div
            ref={containerRef}
            className={`mermaid-diagram ${className} ${isRendering ? 'mermaid-loading' : ''}`}
        />
    );
};

export default MermaidRenderer;
