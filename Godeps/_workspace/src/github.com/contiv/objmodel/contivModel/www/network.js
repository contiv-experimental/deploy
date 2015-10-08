// network.js
// Display Network information

var contivModel = require("../contivModel")

var NetworkPane = React.createClass({
	render: function() {
		var self = this

		if (self.props.networks === undefined) {
			return <div> </div>
		}

        var NetworkSummaryView = contivModel.NetworkSummaryView
        return (
            <div style={{margin: '5%',}}>
                <NetworkSummaryView key="NetworkSummary" networks={self.props.networks}/>
            </div>
        );
	}
});

module.exports = NetworkPane
