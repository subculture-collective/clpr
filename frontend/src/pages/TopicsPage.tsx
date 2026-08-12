import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { CategoryIcon, Container, SEO, Spinner } from '../components';
import { categoryApi } from '../lib/category-api';

export function TopicsPage() {
    const { data, isLoading, isError } = useQuery({
        queryKey: ['categories', 'topic'],
        queryFn: () => categoryApi.listCategories({ type: 'topic', public: true }),
    });
    const topics = data?.categories ?? [];

    return (
        <>
            <SEO
                title='Topics'
                description='Browse Twitch clips by topic, interest, and community.'
                canonicalUrl='/topics'
            />
            <Container className='py-8'>
                <h1 className='text-3xl font-bold text-foreground mb-2'>Topics</h1>
                <p className='text-muted-foreground mb-8'>Find clips across the conversations, interests, and communities you care about.</p>
                {isLoading ? (
                    <div className='flex justify-center py-16'><Spinner size='xl' /></div>
                ) : isError ? (
                    <p className='text-center text-muted-foreground py-16'>Topics could not be loaded.</p>
                ) : topics.length === 0 ? (
                    <p className='text-center text-muted-foreground py-16'>No topics are available yet.</p>
                ) : (
                    <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4'>
                        {topics.map(topic => (
                            <Link key={topic.id} to={`/topics/${topic.slug}`} className='flex items-start gap-3 rounded-xl border border-border bg-surface p-5 hover:border-brand hover:bg-surface-hover transition-colors'>
                                <CategoryIcon icon={topic.icon} size='lg' />
                                <div>
                                    <h2 className='font-semibold text-foreground'>{topic.name}</h2>
                                    {topic.description && <p className='text-sm text-muted-foreground mt-1 line-clamp-2'>{topic.description}</p>}
                                </div>
                            </Link>
                        ))}
                    </div>
                )}
            </Container>
        </>
    );
}
