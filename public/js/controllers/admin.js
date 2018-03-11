define(function () {
	app.registerController('AdminCtrl', ['$scope', '$http', '$sce', '$uibModal', '$rootScope', function ($scope, $http, $sce, $modal, $rootScope) {
		$scope.upgrade = JSON.parse(JSON.stringify($scope.semaphore));
		if ($scope.upgrade && $scope.upgrade.updateBody) {
			$scope.upgrade.updateBody = $sce.trustAsHtml($scope.upgrade.updateBody);
		}

		$scope.checkUpdate = function () {
			$http.get('/upgrade').then(function (response) {
			  var upgrade = response.data;
				if (!upgrade) return;

				if (upgrade.updateBody) {
					upgrade.updateBody = $sce.trustAsHtml(upgrade.updateBody);
				}

				$scope.upgrade = upgrade;
			});
		}

		$scope.doUpgrade = function () {
			var upgradeModal = $modal.open({
				template: '<div class="modal-header"><h3 class="modal-title">Upgrade in progress</h3></div><div class="modal-body"><div class="progress"><div class="progress-bar progress-bar-striped active" style="width: 100%;"></div></div><p ng-if="upgraded">Server has upgraded. It was automatically stopped, if it isn\'t restarted automatically by a process manager please log in and start semaphore again.</p></div>',
				keyboard: false,
				size: 'sm',
				scope: $scope
			});

			$http.post('/upgrade').then(function () {
				$scope.upgraded = true;
				$scope.pollUpgrade(upgradeModal, 0);
			}).catch(function () {
				swal('Error upgrading', arguments, 'error');
			});
		}

		$scope.pollUpgrade = function (modalInstance, attempts) {
			$rootScope.refreshInfo(function (err) {
				if ($rootScope.semaphore.version == $scope.upgrade.update.tag_name.substr(1)) {
					modalInstance.dismiss();
					return;
				}

				setTimeout(function () {
					if (attempts >= 60) {
						swal('Error', 'Upgrade seems to be taking a long time. Check server logs!', 'error');
						modalInstance.dismiss();
						return;
					}

					$scope.pollUpgrade(modalInstance, attempts+1);
				}, 1000);
			});
		}
	}]);
});