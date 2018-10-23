import SeriesList from 'js/components/series-list'
import Layout from 'js/views/layout'
import { Component, h } from 'preact'

export default class SeriesIndex extends Component {

    public render() {

        return <Layout backLink='/'>
            <h1>Series</h1>
            <SeriesList />
        </Layout >
    }

}
