// home.js
// Display Endpoint information

var HomePane = React.createClass({
	render: function() {
		var self = this

		if (self.props.endpoints === undefined) {
			return <div> </div>
		}

		// Walk thru all the endpoints
		var epListView = self.props.endpoints.map(function(ep){
			return (
				<tr key={ep.id} className="info">
					<td>{ep.homingHost}</td>
                    <td>{ep.contName}</td>
                    <td>{ep.netID}</td>
					<td>{ep.ipAddress}</td>
				</tr>
			);
		});

		// Render the pane
		return (
        <div style={{margin: '5%',}}>
			<Table hover>
				<thead>
					<tr>
						<th>Host</th>
                        <th>Container</th>
						<th>Network</th>
						<th>IP address</th>
					</tr>
				</thead>
				<tbody>
			{epListView}
				</tbody>
			</Table>
        </div>
    );
	}
});

module.exports = HomePane
