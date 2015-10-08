// policy.js
// Display Policy information

var contivModel = require("../contivModel")

var PolicyPane = React.createClass({
	render: function() {
		var self = this

		if (self.props.policies === undefined) {
			return <div> </div>
		}

        var PolicySummaryView = contivModel.PolicySummaryView
        return (
            <div style={{margin: '5%',}}>
                <PolicySummaryView key="policySummary" policys={self.props.policies}/>
            </div>
        );
	}
});

module.exports = PolicyPane
