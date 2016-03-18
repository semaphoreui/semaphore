define([
	'services/identities'
], function () {
	app.registerController('IdentitiesCtrl', ['$scope', '$state', 'identities', function($scope, $state, identities) {
		identities.getIdentities(function () {
			$scope.identities = identities.identities;
		});
	}]);
});